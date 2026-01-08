package main

/*
#cgo CFLAGS: -std=c99
#include <stdint.h>
*/
import "C"

import (
	"errors"
	"runtime/cgo"
	"strings"

	"github.com/kadmila/Abyss-Browser/abyss_core/abyst"
	"github.com/kadmila/Abyss-Browser/abyss_core/watchdog"
)

//export AbystClient_Get
func AbystClient_Get(
	h_client C.uintptr_t,
	peer_id_ptr *C.char, peer_id_len C.int,
	path_ptr *C.char, path_len C.int,
	response_handle_out *C.uintptr_t,
) C.uintptr_t {
	client := cgo.Handle(h_client).Value().(*abyst.AbystClient)
	peer_id, peer_id_ok := TryUnmarshalBytes(peer_id_ptr, peer_id_len)
	path, path_ok := TryUnmarshalBytes(path_ptr, path_len)
	if !(peer_id_ok && path_ok) {
		return marshalError(errors.New("nil arguments"))
	}

	resp, err := client.Get(string(peer_id), string(path))
	if err != nil {
		return marshalError(err)
	}

	// Create response handle
	watchdog.CountHandleExport()
	*response_handle_out = C.uintptr_t(cgo.NewHandle(resp))
	return 0
}

//export AbystClient_Post
func AbystClient_Post(
	h_client C.uintptr_t,
	peer_id_ptr *C.char, peer_id_len C.int,
	path_ptr *C.char, path_len C.int,
	content_type_ptr *C.char, content_type_len C.int,
	body_ptr *C.char, body_len C.int,
	response_handle_out *C.uintptr_t,
) C.uintptr_t {
	client := cgo.Handle(h_client).Value().(*abyst.AbystClient)
	peer_id, peer_id_ok := TryUnmarshalBytes(peer_id_ptr, peer_id_len)
	path, path_ok := TryUnmarshalBytes(path_ptr, path_len)
	content_type, content_type_ok := TryUnmarshalBytes(content_type_ptr, content_type_len)
	body, body_ok := TryUnmarshalBytes(body_ptr, body_len)
	if !(peer_id_ok && path_ok && content_type_ok && body_ok) {
		return marshalError(errors.New("nil arguments"))
	}

	resp, err := client.Post(
		string(peer_id),
		string(path),
		string(content_type),
		strings.NewReader(string(body)),
	)
	if err != nil {
		return marshalError(err)
	}

	// Create response handle
	watchdog.CountHandleExport()
	*response_handle_out = C.uintptr_t(cgo.NewHandle(resp))
	return 0
}

//export AbystClient_Head
func AbystClient_Head(
	h_client C.uintptr_t,
	peer_id_ptr *C.char, peer_id_len C.int,
	path_ptr *C.char, path_len C.int,
	response_handle_out *C.uintptr_t,
) C.uintptr_t {
	client := cgo.Handle(h_client).Value().(*abyst.AbystClient)
	peer_id, peer_id_ok := TryUnmarshalBytes(peer_id_ptr, peer_id_len)
	path, path_ok := TryUnmarshalBytes(path_ptr, path_len)
	if !(peer_id_ok && path_ok) {
		return marshalError(errors.New("nil arguments"))
	}

	resp, err := client.Head(string(peer_id), string(path))
	if err != nil {
		return marshalError(err)
	}

	// Create response handle
	watchdog.CountHandleExport()
	*response_handle_out = C.uintptr_t(cgo.NewHandle(resp))
	return 0
}
