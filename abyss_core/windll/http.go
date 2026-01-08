package main

/*
#cgo CFLAGS: -std=c99
#include <stdint.h>
*/
import "C"

import (
	"io"
	"net/http"
	"runtime/cgo"
	"strings"
)

//export HttpResponse_StatusCode
func HttpResponse_StatusCode(h_response C.uintptr_t) C.int {
	resp := cgo.Handle(h_response).Value().(*http.Response)
	return C.int(resp.StatusCode)
}

//export HttpResponse_GetHeader
func HttpResponse_GetHeader(
	h_response C.uintptr_t,
	key_ptr *C.char, key_len C.int,
	value_buf_ptr *C.char, value_buf_len C.int,
) C.int {
	resp := cgo.Handle(h_response).Value().(*http.Response)
	key, ok := TryUnmarshalBytes(key_ptr, key_len)
	if !ok {
		return INVALID_ARGUMENTS
	}

	value := resp.Header.Get(string(key))
	return TryMarshalBytes(value_buf_ptr, value_buf_len, []byte(value))
}

//export HttpResponse_GetAllHeaders
func HttpResponse_GetAllHeaders(
	h_response C.uintptr_t,
	buf_ptr *C.char, buf_len C.int,
) C.int {
	resp := cgo.Handle(h_response).Value().(*http.Response)

	var builder strings.Builder
	for key, values := range resp.Header {
		for _, value := range values {
			builder.WriteString(key)
			builder.WriteString(": ")
			builder.WriteString(value)
			builder.WriteString("\n")
		}
	}

	return TryMarshalBytes(buf_ptr, buf_len, []byte(builder.String()))
}

//export HttpResponse_ReadBody
func HttpResponse_ReadBody(
	h_response C.uintptr_t,
	buf_ptr *C.char, buf_len C.int,
) C.int {
	resp := cgo.Handle(h_response).Value().(*http.Response)
	if resp.Body == nil {
		return 0
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return REMOTE_ERROR
	}

	return TryMarshalBytes(buf_ptr, buf_len, body)
}

//export CloseHttpResponse
func CloseHttpResponse(h_response C.uintptr_t) {
	handle := cgo.Handle(h_response)
	resp := handle.Value().(*http.Response)
	if resp.Body != nil {
		resp.Body.Close()
	}
	deleteHandle(handle)
}
