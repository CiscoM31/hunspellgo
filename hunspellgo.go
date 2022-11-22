package hunspellgo

import (
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
