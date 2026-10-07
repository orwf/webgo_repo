//go:build windows && amd64

package w32

import (
	"sync"
	"syscall"
	"unsafe"
)

// ─────────────────────────────────────────────────────────────────
// WINDOW CONTEXT MAP
// Maps HWND → any user data (the webview instance)
// ─────────────────────────────────────────────────────────────────

var (
	windowContextMu sync.RWMutex
	windowContext   = map[HWND]interface{}{}
)

func SetWindowContext(hwnd HWND, data interface{}) {
	windowContextMu.Lock()
	windowContext[hwnd] = data
	windowContextMu.Unlock()
}

func GetWindowContext(hwnd HWND) interface{} {
	windowContextMu.RLock()
	defer windowContextMu.RUnlock()
	return windowContext[hwnd] // map lookup on missing key returns nil, not panic
}

func DeleteWindowContext(hwnd HWND) {
	windowContextMu.Lock()
	delete(windowContext, hwnd)
	windowContextMu.Unlock()
}

// ─────────────────────────────────────────────────────────────────
// WINDOW CLASS REGISTRATION
// ─────────────────────────────────────────────────────────────────

// RegisterWindowClass registers a Win32 window class with the given
// name and WndProc callback. Returns the class atom.
func RegisterWindowClass(className string, wndProc uintptr) ATOM {
	instance := GetModuleHandle("")

	cn, _ := syscall.UTF16PtrFromString(className)
	wc := WndClassExW{
		Style:      0x0002 | 0x0001, // CS_HREDRAW | CS_VREDRAW
		WndProc:    wndProc,
		Instance:   instance,
		Background: HBRUSH(COLOR_WINDOW + 1),
		ClassName:  cn,
		Cursor:     LoadCursor(0, IDC_ARROW),
	}
	wc.Size = uint32(unsafe.Sizeof(wc))
	return RegisterClassExW(&wc)
}

// ─────────────────────────────────────────────────────────────────
// WINDOW CREATION
// ─────────────────────────────────────────────────────────────────

// CreateMainWindow creates a standard overlapped window.
func CreateMainWindow(className, title string, w, h int32, instance HINSTANCE) HWND {
	return CreateWindowExW(
		0,
		className, title,
		WS_OVERLAPPEDWINDOW|WS_CLIPCHILDREN,
		CW_USEDEFAULT, CW_USEDEFAULT,
		w, h,
		0, 0, instance, 0,
	)
}

func CreateMainWindowS(className, title string, style uint32, w, h int32, instance HINSTANCE) HWND {
	return CreateWindowExW(
		0,
		className, title,
		style,
		CW_USEDEFAULT, CW_USEDEFAULT,
		w, h,
		0, 0, instance, 0,
	)
}

func CreateMainWindowWithStyle(className, title string, w, h int32, hinst HINSTANCE, style uint32) HWND {
	// ... your existing CreateWindowEx logic but passing the 'style' variable ...
	// Make sure it uses style instead of a hardcoded WS_OVERLAPPEDWINDOW
	return CreateMainWindowS(className, title, style, w, h, hinst) // Placeholder
}

// ─────────────────────────────────────────────────────────────────
// MESSAGE PUMP
// ─────────────────────────────────────────────────────────────────

const PM_REMOVE = 0x0001

// RunMessageLoop runs a standard Win32 message loop until WM_QUIT.
func RunMessageLoop(hwnd HWND) {
	var msg Msg
	for {
		r := GetMessage(&msg, 0, 0, 0)
		if r == 0 || r == -1 {
			break
		}
		if !IsDialogMessage(hwnd, &msg) {
			TranslateMessage(&msg)
			DispatchMessage(&msg)
		}
	}
}

// ─────────────────────────────────────────────────────────────────
// DISPATCH QUEUE
// Cross-thread dispatch: post a func() to be run on the UI thread.
// ─────────────────────────────────────────────────────────────────

// dispatchQueue stores pending funcs for the UI thread.
// We encode the func pointer as lParam in a WM_APP_DISPATCH message.
// This lets background goroutines safely call webview methods.

var (
	dispatchMu sync.Mutex

	dispatchQueues = make(map[HWND][]func())
)

func Dispatch(
	hwnd HWND,
	fn func(),
) {
	if hwnd == 0 || fn == nil {
		return
	}

	dispatchMu.Lock()

	dispatchQueues[hwnd] =
		append(
			dispatchQueues[hwnd],
			fn,
		)

	dispatchMu.Unlock()

	PostMessage(
		hwnd,
		WM_APP_DISPATCH,
		0,
		0,
	)
}

func DrainDispatch(hwnd HWND) {
	dispatchMu.Lock()

	fns :=
		dispatchQueues[hwnd]

	delete(
		dispatchQueues,
		hwnd,
	)

	dispatchMu.Unlock()

	for _, fn := range fns {
		if fn != nil {
			fn()
		}
	}
}

func DeleteDispatchQueue(hwnd HWND) {
	dispatchMu.Lock()

	delete(
		dispatchQueues,
		hwnd,
	)

	dispatchMu.Unlock()
}

// UTF16 helper
func UTF16PtrFromString(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}
