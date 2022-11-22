package hunspellgo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

// To test without installing:
// linux CFLAGS: -I${SRCDIR}/hunspell/src
// linux LDFLAGS: -L${SRCDIR}/hunspell/src/hunspell/.libs -lhunspell-1.7 -Wl,-rpath=${SRCDIR}/hunspell/src/hunspell/.libs

// #cgo linux CFLAGS: -I/usr/local/include
// #cgo linux LDFLAGS: -L/usr/local/lib -lhunspell-1.7 -Wl,-rpath=/usr/local/lib
// #cgo darwin LDFLAGS: -lhunspell-1.7 -L/usr/local/lib
// #cgo darwin CFLAGS: -I/usr/local/include
// #cgo freebsd CFLAGS: -I/usr/local/include
// #cgo freebsd LDFLAGS: -L/usr/local/lib -lhunspell-1.3
// #include <stdlib.h>
// #include <stdio.h>
// #include "hunspell/hunspell.h"
// void printcstring(char* s) {
//     printf("CString: %s\n", s);
// }
import "C"

type Hunhandle struct {
	handle *C.Hunhandle
}

func Hunspell(affpath string, dpath string) *Hunhandle {
	affpathcs := C.CString(affpath)
	defer C.free(unsafe.Pointer(affpathcs))
	dpathcs := C.CString(dpath)
	defer C.free(unsafe.Pointer(dpathcs))
	h := &Hunhandle{}
	h.handle = C.Hunspell_create(affpathcs, dpathcs)
	runtime.SetFinalizer(h, func(handle *Hunhandle) {
		C.Hunspell_destroy(handle.handle)
		h.handle = nil
	})
	return h
}

func sptr(p uintptr) *C.char {
	return *(**C.char)(unsafe.Pointer(p))
}

func CStrings(x **C.char, len int) []string {
	var s []string

	p := uintptr(unsafe.Pointer(x))
	for i := 0; i < len; i++ {
		if sptr(p) == nil {
			break
		}
		s = append(s, C.GoString(sptr(p)))
		p += unsafe.Sizeof(uintptr(0))
	}

	return s
}

var lockStdFileDescriptorsSwapping sync.Mutex

// Capture captures stderr and stdout of a given function call.
func Capture(call func()) (output []byte, err error) {
	originalStdErr, originalStdOut := os.Stderr, os.Stdout
	defer func() {
		lockStdFileDescriptorsSwapping.Lock()
		os.Stderr, os.Stdout = originalStdErr, originalStdOut
		lockStdFileDescriptorsSwapping.Unlock()
	}()

	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer func() {
		e := r.Close()
		if e != nil {
			err = e
		}
		if w != nil {
			e = w.Close()
			if err != nil {
				err = e
			}
		}
	}()

	lockStdFileDescriptorsSwapping.Lock()
	os.Stderr, os.Stdout = w, w
	lockStdFileDescriptorsSwapping.Unlock()

	out := make(chan []byte)
	go func() {
		defer func() {
			// If there is a panic in the function call, copying from "r" does not work anymore.
			_ = recover()
		}()
		var b bytes.Buffer
		_, err := io.Copy(&b, r)
		if err != nil {
			panic(err)
		}
		out <- b.Bytes()
	}()

	call()

	err = w.Close()
	if err != nil {
		return nil, err
	}
	w = nil

	return <-out, err
}

// CaptureWithCGo captures stderr and stdout as well as stderr and stdout of C of a given function call.
func CaptureWithCGo(call func()) (output []byte, err error) {
	lockStdFileDescriptorsSwapping.Lock()

	originalStdout, e := syscall.Dup(syscall.Stdout)
	if e != nil {
		lockStdFileDescriptorsSwapping.Unlock()
		return nil, e
	}

	originalStderr, e := syscall.Dup(syscall.Stderr)
	if e != nil {
		lockStdFileDescriptorsSwapping.Unlock()
		return nil, e
	}

	lockStdFileDescriptorsSwapping.Unlock()
	defer func() {
		lockStdFileDescriptorsSwapping.Lock()
		if e := syscall.Dup2(originalStdout, syscall.Stdout); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()
			err = e
		}
		if e := syscall.Close(originalStdout); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()
			err = e
		}
		if e := syscall.Dup2(originalStderr, syscall.Stderr); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()
			err = e
		}
		if e := syscall.Close(originalStderr); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()
			err = e
		}
		lockStdFileDescriptorsSwapping.Unlock()
	}()

	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer func() {
		e := r.Close()
		if e != nil {
			err = e
		}
		if w != nil {
			e = w.Close()
			if err != nil {
				err = e
			}
		}
	}()

	lockStdFileDescriptorsSwapping.Lock()

	if e := syscall.Dup2(int(w.Fd()), syscall.Stdout); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}
	if e := syscall.Dup2(int(w.Fd()), syscall.Stderr); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}

	lockStdFileDescriptorsSwapping.Unlock()

	out := make(chan []byte)
	go func() {
		defer func() {
			// If there is a panic in the function call, copying from "r" does not work anymore.
			_ = recover()
		}()

		var b bytes.Buffer

		_, err := io.Copy(&b, r)
		if err != nil {
			panic(err)
		}

		out <- b.Bytes()
	}()
	call()
	lockStdFileDescriptorsSwapping.Lock()
	C.fflush(C.stdout)
	err = w.Close()
	if err != nil {
		lockStdFileDescriptorsSwapping.Unlock()
		return nil, err
	}
	w = nil
	if e := syscall.Close(syscall.Stdout); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}
	if e := syscall.Close(syscall.Stderr); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()
		return nil, e
	}
	lockStdFileDescriptorsSwapping.Unlock()
	return <-out, err
}

func (handle *Hunhandle) Add(word string) error {
	wordcs := C.CString(word)
	defer C.free(unsafe.Pointer(wordcs))
	var res C.int
	res = C.Hunspell_add(handle.handle, wordcs)
	if int(res) != 0 {
		return errors.New("failed to add word")
	}
	return nil
}

func (handle *Hunhandle) AddDict(path string) error {
	dpathcs := C.CString(path)
	defer C.free(unsafe.Pointer(dpathcs))

	var res C.int
	// output:
	// 0 = additional dictionary slots available,
	// 1 = slots are now full
	stdout, err := CaptureWithCGo(func() {
		res = C.Hunspell_add_dic(handle.handle, dpathcs)
	})
	if err != nil {
		return err
	}
	if int(res) != 0 {
		return errors.New("failed to load dictionary. Slots are full")
	}
	if bytes.Contains(stdout, []byte("error")) {
		return fmt.Errorf("failed to load dictionary: %s", stdout)
	}
	return nil
}

func (handle *Hunhandle) Suggest(word string) []string {
	wordcs := C.CString(word)
	defer C.free(unsafe.Pointer(wordcs))
	var carray **C.char
	var length C.int
	length = C.Hunspell_suggest(handle.handle, &carray, wordcs)

	if int(length) == 1 {
		return []string{}
	}
	c_strings := CStrings(carray, int(length))

	C.Hunspell_free_list(handle.handle, &carray, length)
	return c_strings
}

func (handle *Hunhandle) Stem(word string) []string {
	wordcs := C.CString(word)
	defer C.free(unsafe.Pointer(wordcs))
	var carray **C.char
	var length C.int
	length = C.Hunspell_stem(handle.handle, &carray, wordcs)

	if int(length) == 1 {
		return []string{}
	}
	strings := CStrings(carray, int(length))

	C.Hunspell_free_list(handle.handle, &carray, length)
	return strings
}

func (handle *Hunhandle) Spell(word string) bool {
	wordcs := C.CString(word)
	defer C.free(unsafe.Pointer(wordcs))
	//C.printcstring(wordcs)
	res := C.Hunspell_spell(handle.handle, wordcs)

	if int(res) == 0 {
		return false
	}
	return true
}

func (handle *Hunhandle) Encoding() string {
	return C.GoString(C.Hunspell_get_dic_encoding(handle.handle))
}
