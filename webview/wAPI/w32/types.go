//go:build windows && amd64

package w32

import "unsafe"

// ─────────────────────────────────────────────────────────────────
// HANDLE TYPES
// ─────────────────────────────────────────────────────────────────

type (
	HANDLE    uintptr
	HWND      uintptr
	HINSTANCE uintptr
	HICON     uintptr
	HCURSOR   uintptr
	HBRUSH    uintptr
	HMENU     uintptr
	HMODULE   uintptr
	HDC       uintptr
	WPARAM    uintptr
	LPARAM    uintptr
	LRESULT   uintptr
	ATOM      uint16
)

// ─────────────────────────────────────────────────────────────────
// RECT — used for WebView2 controller bounds
// ─────────────────────────────────────────────────────────────────

type Rect struct {
	Left, Top, Right, Bottom int32
}

// ─────────────────────────────────────────────────────────────────
// POINT
// ─────────────────────────────────────────────────────────────────

type Point struct {
	X, Y int32
}

// ─────────────────────────────────────────────────────────────────
// MSG — Windows message struct for the message pump
// ─────────────────────────────────────────────────────────────────

type Msg struct {
	Hwnd    HWND
	Message uint32
	WParam  WPARAM
	LParam  LPARAM
	Time    uint32
	Pt      Point
	Private uint32
}

// ─────────────────────────────────────────────────────────────────
// WNDCLASSEX — window class registration
// ─────────────────────────────────────────────────────────────────

type WndClassExW struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   HINSTANCE
	Icon       HICON
	Cursor     HCURSOR
	Background HBRUSH
	MenuName   *uint16
	ClassName  *uint16
	IconSmall  HICON
}

// ─────────────────────────────────────────────────────────────────
// MINMAXINFO — used in WM_GETMINMAXINFO
// ─────────────────────────────────────────────────────────────────

type MinMaxInfo struct {
	Reserved     Point
	MaxSize      Point
	MaxPosition  Point
	MinTrackSize Point
	MaxTrackSize Point
}

// ─────────────────────────────────────────────────────────────────
// WINDOW STYLES
// ─────────────────────────────────────────────────────────────────

const (
	WS_POPUP = 0x80000000
)

const (
	WS_OVERLAPPED   = 0x00000000
	WS_CAPTION      = 0x00C00000
	WS_SYSMENU      = 0x00080000
	WS_THICKFRAME   = 0x00040000
	WS_MINIMIZEBOX  = 0x00020000
	WS_MAXIMIZEBOX  = 0x00010000
	WS_VISIBLE      = 0x10000000
	WS_CLIPCHILDREN = 0x02000000
	WS_CLIPSIBLINGS = 0x04000000

	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU |
		WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX

	WS_EX_APPWINDOW  = 0x00040000
	WS_EX_CLIENTEDGE = 0x00000200
)

// Extended window styles & layered attributes
const (
	GWL_EXSTYLE   = -20
	WS_EX_LAYERED = 0x00080000
	LWA_ALPHA     = 0x2
)

// ─────────────────────────────────────────────────────────────────
// WINDOW MESSAGES
// ─────────────────────────────────────────────────────────────────

const (
	WM_DESTROY       = 0x0002
	WM_SIZE          = 0x0005
	WM_CLOSE         = 0x0010
	WM_QUIT          = 0x0012
	WM_GETMINMAXINFO = 0x0024
	WM_APP           = 0x8000

	// Custom messages for cross-thread dispatch
	WM_APP_DISPATCH = WM_APP + 1
)

// ─────────────────────────────────────────────────────────────────
// SHOW WINDOW
// ─────────────────────────────────────────────────────────────────

const (
	SW_SHOW       = 5
	SW_SHOWNORMAL = 1
)

// ─────────────────────────────────────────────────────────────────
// COLOR / BRUSH
// ─────────────────────────────────────────────────────────────────

const (
	COLOR_WINDOW  = 5
	COLOR_BTNFACE = 15
)

// CW_USEDEFAULT for window positioning
const CW_USEDEFAULT = ^int32(0x7fffffff) // -2147483648

// ─────────────────────────────────────────────────────────────────
// HRESULT helpers
// ─────────────────────────────────────────────────────────────────

type HRESULT int32

const (
	S_OK          HRESULT = 0
	S_FALSE       HRESULT = 1
	E_FAIL        HRESULT = -2147467259 // 0x80004005
	E_NOINTERFACE HRESULT = -2147467262 // 0x80004002
)

func (hr HRESULT) Failed() bool    { return hr < 0 }
func (hr HRESULT) Succeeded() bool { return hr >= 0 }

// ─────────────────────────────────────────────────────────────────
// COINIT flags
// ─────────────────────────────────────────────────────────────────

const (
	COINIT_APARTMENTTHREADED = 0x2
	COINIT_MULTITHREADED     = 0x0
)

// ─────────────────────────────────────────────────────────────────
// SIZE_TYPE for WM_SIZE wParam
// ─────────────────────────────────────────────────────────────────

const (
	SIZE_RESTORED  = 0
	SIZE_MINIMIZED = 1
	SIZE_MAXIMIZED = 2
)

const (
	DWMWA_USE_IMMERSIVE_DARK_MODE = 20
)

// Helper: low/high word extraction
func LOWORD(v uintptr) int32 { return int32(v & 0xFFFF) }
func HIWORD(v uintptr) int32 { return int32((v >> 16) & 0xFFFF) }

// Pointer helper for passing Go pointers to syscalls safely
func UnsafePtr(v interface{}) unsafe.Pointer {
	type iface struct{ _, data unsafe.Pointer }
	return (*iface)(unsafe.Pointer(&v)).data
}
