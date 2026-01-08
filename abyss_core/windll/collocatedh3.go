package main

/*
#cgo CFLAGS: -std=c99
#include <stdint.h>
*/
import "C"

import (
	"bytes"
	"errors"
	"net/http"
	"runtime/cgo"

	"github.com/kadmila/Abyss-Browser/abyss_core/watchdog"
)

//export Http3Client_Get
func Http3Client_Get(
	h_client C.uintptr_t,
	url_ptr *C.char, url_len C.int,
	response_handle_out *C.uintptr_t,
) C.uintptr_t {
	client := cgo.Handle(h_client).Value().(*http.Client)
	url, ok := TryUnmarshalBytes(url_ptr, url_len)
	if !ok {
		return marshalError(errors.New("nil arguments"))
	}

	resp, err := client.Get(string(url))
	if err != nil {
		return marshalError(err)
	}

	// Create response handle
	watchdog.CountHandleExport()
	*response_handle_out = C.uintptr_t(cgo.NewHandle(resp))
	return 0
}

//export Http3Client_Post
func Http3Client_Post(
	h_client C.uintptr_t,
	url_ptr *C.char, url_len C.int,
	content_type_ptr *C.char, content_type_len C.int,
	body_ptr *C.char, body_len C.int,
	response_handle_out *C.uintptr_t,
) C.uintptr_t {
	client := cgo.Handle(h_client).Value().(*http.Client)
	url, url_ok := TryUnmarshalBytes(url_ptr, url_len)
	content_type, content_type_ok := TryUnmarshalBytes(content_type_ptr, content_type_len)
	body, body_ok := TryUnmarshalBytes(body_ptr, body_len)
	if !(url_ok && content_type_ok && body_ok) {
		return marshalError(errors.New("nil arguments"))
	}

	resp, err := client.Post(
		string(url),
		string(content_type),
		bytes.NewReader(body),
	)
	if err != nil {
		return marshalError(err)
	}

	// Create response handle
	watchdog.CountHandleExport()
	*response_handle_out = C.uintptr_t(cgo.NewHandle(resp))
	return 0
}

//export Http3Client_Head
func Http3Client_Head(
	h_client C.uintptr_t,
	url_ptr *C.char, url_len C.int,
	response_handle_out *C.uintptr_t,
) C.uintptr_t {
	client := cgo.Handle(h_client).Value().(*http.Client)
	url, ok := TryUnmarshalBytes(url_ptr, url_len)
	if !ok {
		return marshalError(errors.New("nil arguments"))
	}

	resp, err := client.Head(string(url))
	if err != nil {
		return marshalError(err)
	}

	// Create response handle
	watchdog.CountHandleExport()
	*response_handle_out = C.uintptr_t(cgo.NewHandle(resp))
	return 0
}
