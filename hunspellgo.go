package hunspellgo

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
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

// Add adds a word to the dictionary.
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

// AddDict adds a custom dictionary.
func (handle *Hunhandle) AddDict(path string) error {
	dpathcs := C.CString(path)
	defer C.free(unsafe.Pointer(dpathcs))

	var res C.int
	// output:
	// 0 = additional dictionary slots available,
	// 1 = slots are now full
	stdout, err := captureWithCGo(func() {
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
