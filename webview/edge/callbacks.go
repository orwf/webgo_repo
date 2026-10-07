//go:build windows && amd64

package edge

import (
	"fmt"
	"unsafe"

	"webgo_repo-main/webview/wAPI/com"
)

// ─────────────────────────────────────────────────────────────────
// COM CALLBACK PATTERN
//
// WebView2 uses COM callbacks for async operations:
//   CreateCoreWebView2EnvironmentWithOptions(... handler)
//                                                 ↑
//   handler.Invoke(HRESULT, ICoreWebView2Environment*)
//
// To implement a COM interface in Go without CGo we create a struct
// with a vtable we build ourselves:
//
//   vtable[0] = QueryInterface function pointer
//   vtable[1] = AddRef function pointer
//   vtable[2] = Release function pointer
//   vtable[3] = Invoke function pointer
//
// Each function pointer is a windows.NewCallback thunk.
// ─────────────────────────────────────────────────────────────────

// ─────────────────────────────────────────────────────────────────
// ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandler
// ─────────────────────────────────────────────────────────────────

type environmentCompletedHandler struct {
	vtable *environmentCompletedHandlerVTable
}

type environmentCompletedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

type navigationStartingHandler struct {
	vtable *navigationStartingHandlerVTable
}

type navigationStartingHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

// NewNavigationStartingHandler fires fn when a navigation starts.
// fn receives: (args *ICoreWebView2NavigationStartingEventArgs)
func NewNavigationStartingHandler(fn func(*ICoreWebView2NavigationStartingEventArgs)) *navigationStartingHandler {
	h := &navigationStartingHandler{}
	vt := &navigationStartingHandlerVTable{
		QueryInterface: com.NewComProc(func(this, riid, ppvObject uintptr) uintptr {
			return 0x80004002 // E_NOINTERFACE
		}),
		AddRef:  com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Release: com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Invoke: com.NewComProc(func(this, sender, args uintptr) uintptr {
			eArgs := (*ICoreWebView2NavigationStartingEventArgs)(unsafe.Pointer(args))
			fn(eArgs)
			return 0 // S_OK
		}),
	}
	h.vtable = vt
	return h
}

func (h *navigationStartingHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// NewEnvironmentCompletedHandler creates a COM callback that calls fn
// when the WebView2 environment is ready.
// fn receives: (result HRESULT, environment *ICoreWebView2Environment)
func NewEnvironmentCompletedHandler(fn func(uintptr, *ICoreWebView2Environment)) *environmentCompletedHandler {
	h := &environmentCompletedHandler{}
	vt := &environmentCompletedHandlerVTable{
		QueryInterface: com.NewComProc(func(this, riid, ppvObject uintptr) uintptr {
			return 0x80004002 // E_NOINTERFACE
		}),
		AddRef:  com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Release: com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Invoke: com.NewComProc(func(this, result, environment uintptr) uintptr {
			fn(result, (*ICoreWebView2Environment)(unsafe.Pointer(environment)))
			return 0 // S_OK
		}),
	}
	h.vtable = vt
	return h
}

func (h *environmentCompletedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// ─────────────────────────────────────────────────────────────────
// ICoreWebView2CreateCoreWebView2ControllerCompletedHandler
// ─────────────────────────────────────────────────────────────────

type controllerCompletedHandler struct {
	vtable *controllerCompletedHandlerVTable
}

type controllerCompletedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

// NewControllerCompletedHandler creates a COM callback called when
// the WebView2 controller (the window embed) is ready.
func NewControllerCompletedHandler(fn func(uintptr, *ICoreWebView2Controller)) *controllerCompletedHandler {
	h := &controllerCompletedHandler{}
	vt := &controllerCompletedHandlerVTable{
		QueryInterface: com.NewComProc(func(this, riid, ppvObject uintptr) uintptr {
			return 0x80004002
		}),
		AddRef:  com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Release: com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Invoke: com.NewComProc(func(this, result, controller uintptr) uintptr {
			fn(result, (*ICoreWebView2Controller)(unsafe.Pointer(controller)))
			return 0
		}),
	}
	h.vtable = vt
	return h
}

func (h *controllerCompletedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// ─────────────────────────────────────────────────────────────────
// ICoreWebView2WebMessageReceivedEventHandler
// ─────────────────────────────────────────────────────────────────

type webMessageReceivedHandler struct {
	vtable *webMessageReceivedHandlerVTable
}

type webMessageReceivedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

// ICoreWebView2WebMessageReceivedEventArgs — used in the Invoke callback
type ICoreWebView2WebMessageReceivedEventArgs struct{ vtable uintptr }

// TryGetWebMessageAsString extracts the message string from event args.
// vtable index 4 = TryGetWebMessageAsString
func (a *ICoreWebView2WebMessageReceivedEventArgs) TryGetWebMessageAsString() (string, error) {
	var msgPtr *uint16
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(a)), 4,
		uintptr(unsafe.Pointer(&msgPtr)),
	)
	if err := com.CheckHR(r, "TryGetWebMessageAsString"); err != nil {
		return "", err
	}
	return com.UTF16PtrToString(msgPtr), nil
}

// NewWebMessageReceivedHandler creates a COM handler that calls fn
// whenever window.chrome.webview.postMessage(msg) is called in JS.
func NewWebMessageReceivedHandler(fn func(string)) *webMessageReceivedHandler {
	h := &webMessageReceivedHandler{}
	vt := &webMessageReceivedHandlerVTable{
		QueryInterface: com.NewComProc(func(this, riid, ppvObject uintptr) uintptr {
			return 0x80004002
		}),
		AddRef:  com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Release: com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Invoke: com.NewComProc(func(this, sender, args uintptr) uintptr {
			eArgs := (*ICoreWebView2WebMessageReceivedEventArgs)(unsafe.Pointer(args))
			msg, err := eArgs.TryGetWebMessageAsString()
			if err == nil {
				fn(msg)
			}
			return 0
		}),
	}
	h.vtable = vt
	return h
}

func (h *webMessageReceivedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// ─────────────────────────────────────────────────────────────────
// ICoreWebView2NavigationCompletedEventHandler
// ─────────────────────────────────────────────────────────────────

type navigationCompletedHandler struct {
	vtable *navigationCompletedHandlerVTable
}

type navigationCompletedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

// NewNavigationCompletedHandler fires fn when a navigation finishes.
// fn receives: (isSuccess bool)
func NewNavigationCompletedHandler(fn func(uintptr, *ICoreWebView2NavigationCompletedEventArgs)) *navigationCompletedHandler {
	h := &navigationCompletedHandler{}
	vt := &navigationCompletedHandlerVTable{
		QueryInterface: com.NewComProc(func(this, riid, ppvObject uintptr) uintptr {
			return 0x80004002
		}),
		AddRef:  com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Release: com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Invoke: com.NewComProc(func(this, sender, args uintptr) uintptr {
			eArgs := (*ICoreWebView2NavigationCompletedEventArgs)(unsafe.Pointer(args))
			// Get and log error status
			status, err := eArgs.GetWebErrorStatus()
			if err != nil {
				fmt.Println("[Chromium] NavigationCompleted: failed to get error status:", err)
			} else {
				fmt.Printf("[Chromium] NavigationCompleted: WebErrorStatus=%d\n", status)
			}
			fn(sender, eArgs)
			return 0
		}),
	}
	h.vtable = vt
	return h
}

func (h *navigationCompletedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// ─────────────────────────────────────────────────────────────────
// ICoreWebView2ExecuteScriptCompletedHandler
// ─────────────────────────────────────────────────────────────────

type executeScriptCompletedHandler struct {
	vtable *executeScriptCompletedHandlerVTable
}

type executeScriptCompletedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

// NewExecuteScriptCompletedHandler fires fn with the JSON result of ExecuteScript.
func NewExecuteScriptCompletedHandler(fn func(result string)) *executeScriptCompletedHandler {
	h := &executeScriptCompletedHandler{}
	vt := &executeScriptCompletedHandlerVTable{
		QueryInterface: com.NewComProc(func(this, riid, ppvObject uintptr) uintptr {
			return 0x80004002
		}),
		AddRef:  com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Release: com.NewComProc(func(this uintptr) uintptr { return 1 }),
		Invoke: com.NewComProc(func(this, errorCode, resultPtr uintptr) uintptr {
			if int32(errorCode) >= 0 && resultPtr != 0 {
				fn(com.UTF16PtrToString((*uint16)(unsafe.Pointer(resultPtr))))
			}
			return 0
		}),
	}
	h.vtable = vt
	return h
}

func (h *executeScriptCompletedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}
