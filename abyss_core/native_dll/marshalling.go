package main

import "C"
import (
	"os"
	"time"
	"unsafe"
)

func TryMarshalBytes(buf *C.char, buflen C.int, data []byte) C.int {
	if int(buflen) < len(data) {
		return BUFFER_OVERFLOW
	}

	slice := (*[1 << 28]byte)(unsafe.Pointer(buf))[:buflen]
	copy(slice, data)
	return C.int(len(data))
}

func TryUnmarshalBytes(buf *C.char, buflen C.int) ([]byte, bool) {
	if buf == nil || buflen == 0 {
		return []byte{}, false
	}
	return (*[1 << 28]byte)(unsafe.Pointer(buf))[:buflen], true
}

func DebugLog(content string) {
	file, err := os.OpenFile("abyssnet_dll_crash_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	file.WriteString(time.Now().Format("2006-01-02 15:04:05.999999 -0700 MST") + " " + content + "\n")
	file.Close()
}
