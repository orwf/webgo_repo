//go:build windows && amd64

package w32

import (
	"syscall"
	"unsafe"
)

// ─────────────────────────────────────────────────────────────────
// DLL REFERENCES  — lazy-loaded, no CGo needed
// ─────────────────────────────────────────────────────────────────

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
)

// ─────────────────────────────────────────────────────────────────
// user32 procedures
// ─────────────────────────────────────────────────────────────────

var (
	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procGetMessage       = user32.NewProc("GetMessageW")
	procPeekMessage      = user32.NewProc("PeekMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage  = user32.NewProc("DispatchMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procPostMessage      = user32.NewProc("PostMessageW")
	procSendMessage      = user32.NewProc("SendMessageW")
	procDefWindowProc    = user32.NewProc("DefWindowProcW")
	procGetClientRect    = user32.NewProc("GetClientRect")
	procSetWindowText    = user32.NewProc("SetWindowTextW")
	procLoadIcon         = user32.NewProc("LoadIconW")
	procLoadCursor       = user32.NewProc("LoadCursorW")
	procSetWindowPos     = user32.NewProc("SetWindowPos")
	procGetWindowRect    = user32.NewProc("GetWindowRect")
	procScreenToClient   = user32.NewProc("ScreenToClient")
	procSetFocus         = user32.NewProc("SetFocus")
	procIsDialogMessage  = user32.NewProc("IsDialogMessageW")

	// additional procedures added for window manipulation
	procGetWindowLongW             = user32.NewProc("GetWindowLongW")
	procSetWindowLongW             = user32.NewProc("SetWindowLongW")
	procMoveWindow                 = user32.NewProc("MoveWindow")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
)

// ─────────────────────────────────────────────────────────────────
// kernel32 procedures
// ─────────────────────────────────────────────────────────────────

var (
	procGetModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
	procGetCurrentThreadID = kernel32.NewProc("GetCurrentThreadId")
)

// ─────────────────────────────────────────────────────────────────
// ole32 procedures
// ─────────────────────────────────────────────────────────────────

var (
	procCoInitializeEx = ole32.NewProc("CoInitializeEx")
	procCoUninitialize = ole32.NewProc("CoUninitialize")
	procCoTaskMemFree  = ole32.NewProc("CoTaskMemFree")
)

// ─────────────────────────────────────────────────────────────────
// FUNCTION WRAPPERS
// ─────────────────────────────────────────────────────────────────

func CoInitializeEx(reserved uintptr, flags uint32) HRESULT {
	r, _, _ := procCoInitializeEx.Call(reserved, uintptr(flags))
	return HRESULT(r)
}

func CoUninitialize() {
	procCoUninitialize.Call()
}

func CoTaskMemFree(pv uintptr) {
	procCoTaskMemFree.Call(pv)
}

func GetModuleHandle(moduleName string) HINSTANCE {
	if moduleName == "" {
		r, _, _ := procGetModuleHandleW.Call(0)
		return HINSTANCE(r)
	}
	ptr, _ := syscall.UTF16PtrFromString(moduleName)
	r, _, _ := procGetModuleHandleW.Call(uintptr(unsafe.Pointer(ptr)))
	return HINSTANCE(r)
}

func RegisterClassExW(wc *WndClassExW) ATOM {
	r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(wc)))
	return ATOM(r)
}

func CreateWindowExW(
	exStyle uint32, className, windowName string,
	style uint32, x, y, w, h int32,
	parent HWND, menu HMENU, instance HINSTANCE, param uintptr,
) HWND {
	cn, _ := syscall.UTF16PtrFromString(className)
	wn, _ := syscall.UTF16PtrFromString(windowName)
	r, _, _ := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(cn)),
		uintptr(unsafe.Pointer(wn)),
		uintptr(style),
		uintptr(x), uintptr(y),
		uintptr(w), uintptr(h),
		uintptr(parent),
		uintptr(menu),
		uintptr(instance),
		param,
	)
	return HWND(r)
}

func DestroyWindow(hwnd HWND) bool {
	r, _, _ := procDestroyWindow.Call(uintptr(hwnd))
	return r != 0
}

func ShowWindow(hwnd HWND, cmdShow int) bool {
	r, _, _ := procShowWindow.Call(uintptr(hwnd), uintptr(cmdShow))
	return r != 0
}

func UpdateWindow(hwnd HWND) bool {
	r, _, _ := procUpdateWindow.Call(uintptr(hwnd))
	return r != 0
}

func GetMessage(msg *Msg, hwnd HWND, msgFilterMin, msgFilterMax uint32) int {
	r, _, _ := procGetMessage.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	return int(int32(r))
}

func PeekMessage(msg *Msg, hwnd HWND, msgFilterMin, msgFilterMax, removeMsg uint32) bool {
	r, _, _ := procPeekMessage.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
		uintptr(removeMsg),
	)
	return r != 0
}

func TranslateMessage(msg *Msg) bool {
	r, _, _ := procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
	return r != 0
}

func DispatchMessage(msg *Msg) LRESULT {
	r, _, _ := procDispatchMessage.Call(uintptr(unsafe.Pointer(msg)))
	return LRESULT(r)
}

func PostQuitMessage(exitCode int) {
	procPostQuitMessage.Call(uintptr(exitCode))
}

func PostMessage(hwnd HWND, msg uint32, wParam WPARAM, lParam LPARAM) bool {
	r, _, _ := procPostMessage.Call(uintptr(hwnd), uintptr(msg), uintptr(wParam), uintptr(lParam))
	return r != 0
}

func DefWindowProc(hwnd HWND, msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	r, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(msg), uintptr(wParam), uintptr(lParam))
	return LRESULT(r)
}

func GetClientRect(hwnd HWND, rect *Rect) bool {
	r, _, _ := procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(rect)))
	return r != 0
}

// GetWindowLong retrieves information about the specified window.
func GetWindowLong(hwnd HWND, index int) int {
	r, _, _ := procGetWindowLongW.Call(uintptr(hwnd), uintptr(index))
	return int(r)
}

// SetWindowLong changes an attribute of the specified window.
func SetWindowLong(hwnd HWND, index int, value int) int {
	r, _, _ := procSetWindowLongW.Call(uintptr(hwnd), uintptr(index), uintptr(value))
	return int(r)
}

// MoveWindow changes the position and size of the specified window.
func MoveWindow(hwnd HWND, x, y, w, h int32, repaint bool) bool {
	rp := uintptr(0)
	if repaint {
		rp = 1
	}
	r, _, _ := procMoveWindow.Call(
		uintptr(hwnd),
		uintptr(x), uintptr(y),
		uintptr(w), uintptr(h),
		rp,
	)
	return r != 0
}

// ScreenToClient converts the screen coordinates of a specified point on the screen to client-area coordinates.
func ScreenToClient(hwnd HWND, lpPoint *Point) bool {
	ret, _, _ := procScreenToClient.Call(uintptr(hwnd), uintptr(unsafe.Pointer(lpPoint)))
	return ret != 0
}

// SetLayeredWindowAttributes applies transparency settings to a layered window.
func SetLayeredWindowAttributes(hwnd HWND, crKey uint32, bAlpha byte, dwFlags uint32) bool {
	r, _, _ := procSetLayeredWindowAttributes.Call(
		uintptr(hwnd),
		uintptr(crKey),
		uintptr(bAlpha),
		uintptr(dwFlags),
	)
	return r != 0
}

func SetWindowText(hwnd HWND, text string) bool {
	ptr, _ := syscall.UTF16PtrFromString(text)
	r, _, _ := procSetWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ptr)))
	return r != 0
}

func LoadCursor(instance HINSTANCE, cursorName uintptr) HCURSOR {
	r, _, _ := procLoadCursor.Call(uintptr(instance), cursorName)
	return HCURSOR(r)
}

// IDC_ARROW standard cursor
const IDC_ARROW = uintptr(32512)

func SetFocus(hwnd HWND) HWND {
	r, _, _ := procSetFocus.Call(uintptr(hwnd))
	return HWND(r)
}

func IsDialogMessage(hwnd HWND, msg *Msg) bool {
	r, _, _ := procIsDialogMessage.Call(uintptr(hwnd), uintptr(unsafe.Pointer(msg)))
	return r != 0
}
