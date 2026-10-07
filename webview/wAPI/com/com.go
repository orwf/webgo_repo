//go:build windows && amd64

// Package com provides pure-Go COM interop for amd64 Windows.
//
// COM (Component Object Model) works by having each object expose a pointer
// to a vtable — an array of function pointers. The first argument to every
// COM method is "this" (the object pointer itself).
//
// On amd64 Windows the calling convention is __fastcall:
//
//	rcx = arg0 (this)
//	rdx = arg1
//	r8  = arg2
//	r9  = arg3
//	stack for arg4+
//
// Go's syscall.SyscallN maps directly to this, so COM calls work with no CGo.
package com

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ─────────────────────────────────────────────────────────────────
// GUID
// ─────────────────────────────────────────────────────────────────

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	ole32             = syscall.NewLazyDLL("ole32.dll")
	procCoTaskMemFree = ole32.NewProc("CoTaskMemFree")
)

func CoTaskMemFree(ptr uintptr) {
	if ptr == 0 {
		return
	}

	procCoTaskMemFree.Call(ptr)
}

func GUIDEqual(a, b *GUID) bool {
	if a == nil || b == nil {
		return false
	}

	if a.Data1 != b.Data1 ||
		a.Data2 != b.Data2 ||
		a.Data3 != b.Data3 {
		return false
	}

	return a.Data4 == b.Data4
}

// NewGUID parses a GUID string like "{F3B30B28-...}"
func NewGUID(s string) *GUID {
	// Strip braces
	if len(s) > 0 && s[0] == '{' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '}' {
		s = s[:len(s)-1]
	}
	var g GUID
	fmt.Sscanf(s, "%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		&g.Data1, &g.Data2, &g.Data3,
		&g.Data4[0], &g.Data4[1],
		&g.Data4[2], &g.Data4[3], &g.Data4[4],
		&g.Data4[5], &g.Data4[6], &g.Data4[7],
	)
	return &g
}

func (g *GUID) String() string {
	return fmt.Sprintf("{%08X-%04X-%04X-%02X%02X-%02X%02X%02X%02X%02X%02X}",
		g.Data1, g.Data2, g.Data3,
		g.Data4[0], g.Data4[1],
		g.Data4[2], g.Data4[3], g.Data4[4],
		g.Data4[5], g.Data4[6], g.Data4[7],
	)
}

// ─────────────────────────────────────────────────────────────────
// VTABLE CALL — the core of no-CGo COM
// ─────────────────────────────────────────────────────────────────

// VTableCall calls the COM method at vtable index `idx` on object `this`.
// Additional arguments are passed as uintptr values.
//
// Layout in memory:
//
//	this → [vtable*]
//	vtable → [method0, method1, ..., methodN]
//	vtable[idx](this, args...)
func VTableCall(this uintptr, idx int, args ...uintptr) (uintptr, uintptr, error) {
	// Dereference: this → vtable pointer → method pointer at index
	vtable := *(*uintptr)(unsafe.Pointer(this))
	method := *(*uintptr)(unsafe.Pointer(vtable + uintptr(idx)*unsafe.Sizeof(uintptr(0))))

	// Build arg slice: prepend 'this'
	callArgs := make([]uintptr, 0, len(args)+1)
	callArgs = append(callArgs, this)
	callArgs = append(callArgs, args...)

	return syscall.SyscallN(method, callArgs...)
}

// ─────────────────────────────────────────────────────────────────
// COMPROC — Go func → COM callback (no assembly, uses windows.NewCallback)
// ─────────────────────────────────────────────────────────────────

// ComProc is a COM-callable procedure wrapping a Go function.
// windows.NewCallback creates an amd64 thunk at runtime.
type ComProc uintptr

// NewComProc wraps fn (must be a func with uintptr args/return)
// into a COM-callable thunk. The result is a stable function pointer
// safe to pass to COM interfaces.
func NewComProc(fn interface{}) ComProc {
	return ComProc(windows.NewCallback(fn))
}

// Call invokes the ComProc as if it were a COM method.
func (p ComProc) Call(args ...uintptr) (uintptr, uintptr, error) {
	return syscall.SyscallN(uintptr(p), args...)
}

// ─────────────────────────────────────────────────────────────────
// IUNKNOWN — base COM interface (all COM objects implement this)
// ─────────────────────────────────────────────────────────────────

// IUnknownVTable mirrors the COM IUnknown vtable:
//
//	index 0: QueryInterface
//	index 1: AddRef
//	index 2: Release
type IUnknown struct {
	vtable uintptr // pointer to vtable
}

func (u *IUnknown) AddRef() uintptr {
	r, _, _ := VTableCall(uintptr(unsafe.Pointer(u)), 1)
	return r
}

func (u *IUnknown) Release() uintptr {
	r, _, _ := VTableCall(uintptr(unsafe.Pointer(u)), 2)
	return r
}

// ─────────────────────────────────────────────────────────────────
// HRESULT helpers
// ─────────────────────────────────────────────────────────────────

type HRESULT = int32

const (
	S_OK          HRESULT = 0
	S_FALSE       HRESULT = 1
	E_FAIL        uintptr = 0x80004005
	E_NOINTERFACE uintptr = 0x80004002
	E_POINTER     uintptr = 0x80004003
)

func Failed(hr uintptr) bool    { return int32(hr) < 0 }
func Succeeded(hr uintptr) bool { return int32(hr) >= 0 }

func CheckHR(hr uintptr, op string) error {
	if int32(hr) < 0 {
		return fmt.Errorf("%s failed: HRESULT 0x%08X", op, uint32(hr))
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────
// UTF16 helpers used by COM string params
// ─────────────────────────────────────────────────────────────────

func UTF16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	return windows.UTF16PtrToString(p)
}
