//go:build windows && amd64

package edge

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/orwf/webgo_repo/webview/wAPI/com"
)

var (
	iidIUnknown = com.NewGUID("{00000000-0000-0000-C000-000000000046}")

	iidEnvironmentCompleted = com.NewGUID("{4E8A3389-C9D8-4BD2-B6B5-124FEE6CC14D}")

	iidControllerCompleted = com.NewGUID("{6C4819F3-C9B7-4260-8127-C9F5BDE7F68C}")

	iidNavigationStarting = com.NewGUID("{9ADBE429-F36D-432B-9DDC-F8881FBD76E3}")

	iidNavigationCompleted = com.NewGUID("{D33A35BF-1C49-4F98-93AB-006E0533FE1C}")

	iidWebMessageReceived = com.NewGUID("{57213F19-00E6-49FA-8E07-898EA01ECBD2}")

	iidExecuteScriptCompleted = com.NewGUID("{49511172-CC67-4BCA-9923-137112F4C4CC}")

	iidProcessFailed = com.NewGUID("{79E0AEA4-990B-42D9-AA1D-0FCC2E5BC7F1}")
)

func callbackQueryInterface(
	this uintptr,
	riid uintptr,
	ppvObject uintptr,
	interfaceIID *com.GUID,
	refs *uint32,
) uintptr {
	if ppvObject == 0 {
		return uintptr(uint32(com.E_POINTER))
	}

	*(*uintptr)(unsafe.Pointer(ppvObject)) = 0

	if riid == 0 {
		return uintptr(uint32(com.E_NOINTERFACE))
	}

	requested := (*com.GUID)(unsafe.Pointer(riid))

	if !com.GUIDEqual(requested, iidIUnknown) &&
		!com.GUIDEqual(requested, interfaceIID) {

		return uintptr(uint32(com.E_NOINTERFACE))
	}

	*(*uintptr)(unsafe.Pointer(ppvObject)) = this

	atomic.AddUint32(refs, 1)

	return uintptr(uint32(com.S_OK))
}

func callbackAddRef(refs *uint32) uintptr {
	return uintptr(atomic.AddUint32(refs, 1))
}

func callbackRelease(refs *uint32) uintptr {
	for {
		old := atomic.LoadUint32(refs)

		if old == 0 {
			return 0
		}

		next := old - 1

		if atomic.CompareAndSwapUint32(
			refs,
			old,
			next,
		) {
			return uintptr(next)
		}
	}
}

func safeCallback(fn func()) (hr uintptr) {
	hr = uintptr(uint32(com.S_OK))

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(
				"[WebGo] panic inside COM callback:",
				r,
			)

			hr = uintptr(uint32(com.E_FAIL))
		}
	}()

	fn()

	return
}

// ────────────────────────────────────────────────────────────────
// Environment completed
// ────────────────────────────────────────────────────────────────

type environmentCompletedHandler struct {
	vtable *environmentCompletedHandlerVTable
	refs   uint32
}

type environmentCompletedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

func NewEnvironmentCompletedHandler(
	fn func(uintptr, *ICoreWebView2Environment),
) *environmentCompletedHandler {

	h := &environmentCompletedHandler{
		refs: 1,
	}

	h.vtable = &environmentCompletedHandlerVTable{
		QueryInterface: com.NewComProc(
			func(this, riid, ppvObject uintptr) uintptr {
				obj := (*environmentCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackQueryInterface(
					this,
					riid,
					ppvObject,
					iidEnvironmentCompleted,
					&obj.refs,
				)
			},
		),

		AddRef: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*environmentCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackAddRef(&obj.refs)
			},
		),

		Release: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*environmentCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackRelease(&obj.refs)
			},
		),

		Invoke: com.NewComProc(
			func(
				this,
				result,
				environment uintptr,
			) uintptr {
				return safeCallback(func() {
					fn(
						result,
						(*ICoreWebView2Environment)(
							unsafe.Pointer(environment),
						),
					)
				})
			},
		),
	}

	return h
}

func (h *environmentCompletedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// ────────────────────────────────────────────────────────────────
// Controller completed
// ────────────────────────────────────────────────────────────────

type controllerCompletedHandler struct {
	vtable *controllerCompletedHandlerVTable
	refs   uint32
}

type controllerCompletedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

func NewControllerCompletedHandler(
	fn func(uintptr, *ICoreWebView2Controller),
) *controllerCompletedHandler {

	h := &controllerCompletedHandler{
		refs: 1,
	}

	h.vtable = &controllerCompletedHandlerVTable{
		QueryInterface: com.NewComProc(
			func(this, riid, ppvObject uintptr) uintptr {
				obj := (*controllerCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackQueryInterface(
					this,
					riid,
					ppvObject,
					iidControllerCompleted,
					&obj.refs,
				)
			},
		),

		AddRef: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*controllerCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackAddRef(&obj.refs)
			},
		),

		Release: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*controllerCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackRelease(&obj.refs)
			},
		),

		Invoke: com.NewComProc(
			func(
				this,
				result,
				controller uintptr,
			) uintptr {
				return safeCallback(func() {
					fn(
						result,
						(*ICoreWebView2Controller)(
							unsafe.Pointer(controller),
						),
					)
				})
			},
		),
	}

	return h
}

func (h *controllerCompletedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// ────────────────────────────────────────────────────────────────
// Navigation starting
// ────────────────────────────────────────────────────────────────

type navigationStartingHandler struct {
	vtable *navigationStartingHandlerVTable
	refs   uint32
}

type navigationStartingHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

func NewNavigationStartingHandler(
	fn func(*ICoreWebView2NavigationStartingEventArgs),
) *navigationStartingHandler {

	h := &navigationStartingHandler{
		refs: 1,
	}

	h.vtable = &navigationStartingHandlerVTable{
		QueryInterface: com.NewComProc(
			func(this, riid, ppvObject uintptr) uintptr {
				obj := (*navigationStartingHandler)(
					unsafe.Pointer(this),
				)

				return callbackQueryInterface(
					this,
					riid,
					ppvObject,
					iidNavigationStarting,
					&obj.refs,
				)
			},
		),

		AddRef: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*navigationStartingHandler)(
					unsafe.Pointer(this),
				)

				return callbackAddRef(&obj.refs)
			},
		),

		Release: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*navigationStartingHandler)(
					unsafe.Pointer(this),
				)

				return callbackRelease(&obj.refs)
			},
		),

		Invoke: com.NewComProc(
			func(
				this,
				sender,
				args uintptr,
			) uintptr {
				return safeCallback(func() {
					fn(
						(*ICoreWebView2NavigationStartingEventArgs)(
							unsafe.Pointer(args),
						),
					)
				})
			},
		),
	}

	return h
}

func (h *navigationStartingHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// ────────────────────────────────────────────────────────────────
// Web message received
// ────────────────────────────────────────────────────────────────

type webMessageReceivedHandler struct {
	vtable *webMessageReceivedHandlerVTable
	refs   uint32
}

type webMessageReceivedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

type ICoreWebView2WebMessageReceivedEventArgs struct {
	vtable uintptr
}

func (
	a *ICoreWebView2WebMessageReceivedEventArgs,
) TryGetWebMessageAsString() (string, error) {

	var msgPtr *uint16

	r, _, _ := com.VTableCall(
		uintptr(unsafe.Pointer(a)),
		4,
		uintptr(unsafe.Pointer(&msgPtr)),
	)

	if err := com.CheckHR(
		r,
		"TryGetWebMessageAsString",
	); err != nil {
		return "", err
	}

	if msgPtr == nil {
		return "", nil
	}

	msg := com.UTF16PtrToString(msgPtr)

	com.CoTaskMemFree(
		uintptr(unsafe.Pointer(msgPtr)),
	)

	return msg, nil
}

func NewWebMessageReceivedHandler(
	fn func(string),
) *webMessageReceivedHandler {

	h := &webMessageReceivedHandler{
		refs: 1,
	}

	h.vtable = &webMessageReceivedHandlerVTable{
		QueryInterface: com.NewComProc(
			func(this, riid, ppvObject uintptr) uintptr {
				obj := (*webMessageReceivedHandler)(
					unsafe.Pointer(this),
				)

				return callbackQueryInterface(
					this,
					riid,
					ppvObject,
					iidWebMessageReceived,
					&obj.refs,
				)
			},
		),

		AddRef: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*webMessageReceivedHandler)(
					unsafe.Pointer(this),
				)

				return callbackAddRef(&obj.refs)
			},
		),

		Release: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*webMessageReceivedHandler)(
					unsafe.Pointer(this),
				)

				return callbackRelease(&obj.refs)
			},
		),

		Invoke: com.NewComProc(
			func(
				this,
				sender,
				args uintptr,
			) uintptr {
				return safeCallback(func() {
					e := (*ICoreWebView2WebMessageReceivedEventArgs)(
						unsafe.Pointer(args),
					)

					msg, err :=
						e.TryGetWebMessageAsString()

					if err == nil {
						fn(msg)
					}
				})
			},
		),
	}

	return h
}

func (h *webMessageReceivedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// ────────────────────────────────────────────────────────────────
// Navigation completed
// ────────────────────────────────────────────────────────────────

type navigationCompletedHandler struct {
	vtable *navigationCompletedHandlerVTable
	refs   uint32
}

type navigationCompletedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

func NewNavigationCompletedHandler(
	fn func(
		uintptr,
		*ICoreWebView2NavigationCompletedEventArgs,
	),
) *navigationCompletedHandler {

	h := &navigationCompletedHandler{
		refs: 1,
	}

	h.vtable = &navigationCompletedHandlerVTable{
		QueryInterface: com.NewComProc(
			func(this, riid, ppvObject uintptr) uintptr {
				obj := (*navigationCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackQueryInterface(
					this,
					riid,
					ppvObject,
					iidNavigationCompleted,
					&obj.refs,
				)
			},
		),

		AddRef: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*navigationCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackAddRef(&obj.refs)
			},
		),

		Release: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*navigationCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackRelease(&obj.refs)
			},
		),

		Invoke: com.NewComProc(
			func(
				this,
				sender,
				args uintptr,
			) uintptr {

				return safeCallback(func() {
					e :=
						(*ICoreWebView2NavigationCompletedEventArgs)(
							unsafe.Pointer(args),
						)

					fn(sender, e)
				})
			},
		),
	}

	return h
}

func (h *navigationCompletedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}

// ────────────────────────────────────────────────────────────────
// ExecuteScript completed
// ────────────────────────────────────────────────────────────────

type executeScriptCompletedHandler struct {
	vtable *executeScriptCompletedHandlerVTable
	refs   uint32
}

type executeScriptCompletedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef         com.ComProc
	Release        com.ComProc
	Invoke         com.ComProc
}

func NewExecuteScriptCompletedHandler(
	fn func(string),
) *executeScriptCompletedHandler {

	h := &executeScriptCompletedHandler{
		refs: 1,
	}

	h.vtable = &executeScriptCompletedHandlerVTable{
		QueryInterface: com.NewComProc(
			func(this, riid, ppvObject uintptr) uintptr {
				obj := (*executeScriptCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackQueryInterface(
					this,
					riid,
					ppvObject,
					iidExecuteScriptCompleted,
					&obj.refs,
				)
			},
		),

		AddRef: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*executeScriptCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackAddRef(&obj.refs)
			},
		),

		Release: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*executeScriptCompletedHandler)(
					unsafe.Pointer(this),
				)

				return callbackRelease(&obj.refs)
			},
		),

		Invoke: com.NewComProc(
			func(
				this,
				errorCode,
				resultPtr uintptr,
			) uintptr {

				return safeCallback(func() {
					if int32(errorCode) < 0 {
						return
					}

					if resultPtr == 0 {
						fn("")
						return
					}

					fn(
						com.UTF16PtrToString(
							(*uint16)(
								unsafe.Pointer(resultPtr),
							),
						),
					)
				})
			},
		),
	}

	return h
}

func (h *executeScriptCompletedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}


// ────────────────────────────────────────────────────────────────
// Process failed
// ────────────────────────────────────────────────────────────────

type processFailedHandler struct {
	vtable *processFailedHandlerVTable
	refs   uint32
}

type processFailedHandlerVTable struct {
	QueryInterface com.ComProc
	AddRef          com.ComProc
	Release         com.ComProc
	Invoke          com.ComProc
}

func NewProcessFailedHandler(
	fn func(*ICoreWebView2ProcessFailedEventArgs),
) *processFailedHandler {
	h := &processFailedHandler{refs: 1}

	h.vtable = &processFailedHandlerVTable{
		QueryInterface: com.NewComProc(
			func(this, riid, ppvObject uintptr) uintptr {
				obj := (*processFailedHandler)(unsafe.Pointer(this))
				return callbackQueryInterface(
					this,
					riid,
					ppvObject,
					iidProcessFailed,
					&obj.refs,
				)
			},
		),
		AddRef: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*processFailedHandler)(unsafe.Pointer(this))
				return callbackAddRef(&obj.refs)
			},
		),
		Release: com.NewComProc(
			func(this uintptr) uintptr {
				obj := (*processFailedHandler)(unsafe.Pointer(this))
				return callbackRelease(&obj.refs)
			},
		),
		Invoke: com.NewComProc(
			func(this, sender, args uintptr) uintptr {
				return safeCallback(func() {
					fn((*ICoreWebView2ProcessFailedEventArgs)(
						unsafe.Pointer(args),
					))
				})
			},
		),
	}

	return h
}

func (h *processFailedHandler) AsPtr() uintptr {
	return uintptr(unsafe.Pointer(h))
}
