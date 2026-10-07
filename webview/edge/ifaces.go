//go:build windows && amd64

// Package edge provides the ICoreWebView2* COM interface bindings.
//
// Every struct here mirrors a COM interface. The pattern is:
//
//	type IFoo struct { vtable uintptr }
//
//	func (i *IFoo) SomeMethod(arg string) error {
//	    ptr, _ := syscall.UTF16PtrFromString(arg)
//	    r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(i)), vtableIndex, uintptr(unsafe.Pointer(ptr)))
//	    return com.CheckHR(r, "SomeMethod")
//	}
//
// Vtable indices come from the official WebView2 IDL/headers.
// IUnknown always occupies indices 0 (QueryInterface), 1 (AddRef), 2 (Release).
package edge

import (
	"fmt"
	"syscall"
	"unsafe"

	"webgo_repo-main/webview/wAPI/com"
)

// ─────────────────────────────────────────────────────────────────
// ICoreWebView2Environment
// vtable indices (after IUnknown: 0,1,2):
//   3  CreateCoreWebView2Controller
//   4  CreateWebResourceResponse
//   5  get_BrowserVersionString
//   6  add_NewBrowserVersionAvailable
//   7  remove_NewBrowserVersionAvailable
// ─────────────────────────────────────────────────────────────────

type ICoreWebView2Environment struct{ vtable uintptr }
type ICoreWebView2NavigationStartingEventArgs struct{ vtable uintptr }
type ICoreWebView2NavigationCompletedEventArgs struct{ vtable uintptr }

func (a *ICoreWebView2NavigationCompletedEventArgs) GetWebErrorStatus() (uint32, error) {
	var status uint32
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(a)), 5,
		uintptr(unsafe.Pointer(&status)),
	)
	return status, com.CheckHR(r, "get_WebErrorStatus")
}
func (a *ICoreWebView2NavigationStartingEventArgs) GetUri() (string, error) {
	var uriPtr *uint16
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(a)), 3,
		uintptr(unsafe.Pointer(&uriPtr)),
	)
	if err := com.CheckHR(r, "get_Uri"); err != nil {
		return "", err
	}
	return com.UTF16PtrToString(uriPtr), nil
}

func (w *ICoreWebView2) AddNavigationStartingHandler(handler uintptr) (EventRegistrationToken, error) {
	var token EventRegistrationToken
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 7,
		handler,
		uintptr(unsafe.Pointer(&token)),
	)
	err := com.CheckHR(r, "add_NavigationStarting")
	fmt.Printf("[ifaces] AddNavigationStartingHandler HRESULT=0x%08X token=%v err=%v\n", uint32(r), token, err)
	return token, err
}

// CreateCoreWebView2Controller creates the webview controller embedded in parentHWND.
// completedHandler is an ICoreWebView2CreateCoreWebView2ControllerCompletedHandler.
func (e *ICoreWebView2Environment) CreateCoreWebView2Controller(
	parentHWND uintptr,
	completedHandler uintptr,
) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(e)), 3,
		parentHWND,
		completedHandler,
	)
	return com.CheckHR(r, "CreateCoreWebView2Controller")
}

func (e *ICoreWebView2Environment) GetBrowserVersionString() (string, error) {
	var versionPtr *uint16
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(e)), 5,
		uintptr(unsafe.Pointer(&versionPtr)),
	)
	if err := com.CheckHR(r, "get_BrowserVersionString"); err != nil {
		return "", err
	}
	ver := com.UTF16PtrToString(versionPtr)
	return ver, nil
}

func (e *ICoreWebView2Environment) AddRef() uintptr {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(e)), 1)
	return r
}

func (e *ICoreWebView2Environment) Release() uintptr {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(e)), 2)
	return r
}

// ─────────────────────────────────────────────────────────────────
// ICoreWebView2Controller
// vtable indices (after IUnknown: 0,1,2):
//   3  get_IsVisible
//   4  put_IsVisible
//   5  get_Bounds
//   6  put_Bounds
//   7  get_ZoomFactor
//   8  put_ZoomFactor
//   9  add_ZoomFactorChanged
//   10 remove_ZoomFactorChanged
//   11 SetBoundsAndZoomFactor
//   12 MoveFocus
//   13 add_MoveFocusRequested
//   14 remove_MoveFocusRequested
//   15 add_GotFocus
//   16 remove_GotFocus
//   17 add_LostFocus
//   18 remove_LostFocus
//   19 add_AcceleratorKeyPressed
//   20 remove_AcceleratorKeyPressed
//   21 get_ParentWindow
//   22 put_ParentWindow
//   23 NotifyParentWindowPositionChanged
//   24 Close
//   25 get_CoreWebView2
// ─────────────────────────────────────────────────────────────────

type ICoreWebView2Controller struct{ vtable uintptr }

func (c *ICoreWebView2Controller) AddRef() uintptr {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(c)), 1)
	return r
}

func (c *ICoreWebView2Controller) Release() uintptr {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(c)), 2)
	return r
}

func (c *ICoreWebView2Controller) PutIsVisible(visible bool) error {
	v := uintptr(0)
	if visible {
		v = 1
	}
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(c)), 4, v)
	return com.CheckHR(r, "put_IsVisible")
}

func (c *ICoreWebView2Controller) GetBounds() (*Rect, error) {
	var rect Rect
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(c)), 5,
		uintptr(unsafe.Pointer(&rect)),
	)
	return &rect, com.CheckHR(r, "get_Bounds")
}

func (c *ICoreWebView2Controller) PutBounds(rect Rect) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(c)), 6,
		uintptr(unsafe.Pointer(&rect)),
	)
	return com.CheckHR(r, "put_Bounds")
}

func (c *ICoreWebView2Controller) MoveFocus(reason uintptr) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(c)), 12, reason)
	return com.CheckHR(r, "MoveFocus")
}

func (c *ICoreWebView2Controller) NotifyParentWindowPositionChanged() error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(c)), 23)
	return com.CheckHR(r, "NotifyParentWindowPositionChanged")
}

func (c *ICoreWebView2Controller) Close() error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(c)), 24)
	return com.CheckHR(r, "Close")
}

// GetICoreWebView2 returns the ICoreWebView2 for this controller.
func (c *ICoreWebView2Controller) GetICoreWebView2() (*ICoreWebView2, error) {
	var webview *ICoreWebView2
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(c)), 25,
		uintptr(unsafe.Pointer(&webview)),
	)
	return webview, com.CheckHR(r, "get_CoreWebView2")
}

// ─────────────────────────────────────────────────────────────────
// ICoreWebView2
// vtable indices (after IUnknown: 0,1,2):
//   3  get_Settings
//   4  get_Source
//   5  Navigate
//   6  NavigateToString
//   7  add_NavigationStarting
//   8  remove_NavigationStarting
//   9  add_ContentLoading
//  10  remove_ContentLoading
//  11  add_SourceChanged
//  12  remove_SourceChanged
//  13  add_HistoryChanged
//  14  remove_HistoryChanged
//  15  add_NavigationCompleted
//  16  remove_NavigationCompleted
//  17  add_FrameNavigationStarting
//  18  remove_FrameNavigationStarting
//  19  add_FrameNavigationCompleted
//  20  remove_FrameNavigationCompleted
//  21  add_ScriptDialogOpening
//  22  remove_ScriptDialogOpening
//  23  add_PermissionRequested
//  24  remove_PermissionRequested
//  25  add_ProcessFailed
//  26  remove_ProcessFailed
//  27  AddScriptToExecuteOnDocumentCreated
//  28  RemoveScriptToExecuteOnDocumentCreated
//  29  ExecuteScript
//  30  CapturePreview
//  31  Reload
//  32  PostWebMessageAsJson
//  33  PostWebMessageAsString
//  34  add_WebMessageReceived
//  35  remove_WebMessageReceived
//  36  CallDevToolsProtocolMethod
//  37  get_BrowserProcessId
//  38  get_CanGoBack
//  39  get_CanGoForward
//  40  GoBack
//  41  GoForward
//  42  GetDevToolsProtocolEventReceiver
//  43  Stop
//  44  add_NewWindowRequested
//  45  remove_NewWindowRequested
//  46  add_DocumentTitleChanged
//  47  remove_DocumentTitleChanged
//  48  get_DocumentTitle
//  49  AddHostObjectToScript
//  50  RemoveHostObjectFromScript
//  51  OpenDevToolsWindow
//  52  add_ContainsFullScreenElementChanged
//  53  remove_ContainsFullScreenElementChanged
//  54  get_ContainsFullScreenElement
//  55  add_WebResourceRequested
//  56  remove_WebResourceRequested
//  57  AddWebResourceRequestedFilter
//  58  RemoveWebResourceRequestedFilter
//  59  add_WindowCloseRequested
//  60  remove_WindowCloseRequested
// ─────────────────────────────────────────────────────────────────

type ICoreWebView2 struct{ vtable uintptr }

func (w *ICoreWebView2) AddRef() uintptr {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 1)
	return r
}

func (w *ICoreWebView2) Release() uintptr {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 2)
	return r
}

func (w *ICoreWebView2) GetSettings() (*ICoreWebView2Settings, error) {
	var settings *ICoreWebView2Settings
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 3,
		uintptr(unsafe.Pointer(&settings)),
	)
	return settings, com.CheckHR(r, "get_Settings")
}

func (w *ICoreWebView2) Navigate(url string) error {
	ptr, _ := syscall.UTF16PtrFromString(url)
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 5,
		uintptr(unsafe.Pointer(ptr)),
	)
	return com.CheckHR(r, "Navigate")
}

func (w *ICoreWebView2) NavigateToString(html string) error {
	ptr, _ := syscall.UTF16PtrFromString(html)
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 6,
		uintptr(unsafe.Pointer(ptr)),
	)
	return com.CheckHR(r, "NavigateToString")
}

func (w *ICoreWebView2) AddScriptToExecuteOnDocumentCreated(script string, handler uintptr) error {
	ptr, _ := syscall.UTF16PtrFromString(script)
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 27,
		uintptr(unsafe.Pointer(ptr)),
		handler,
	)
	return com.CheckHR(r, "AddScriptToExecuteOnDocumentCreated")
}

func (w *ICoreWebView2) ExecuteScript(script string, handler uintptr) error {
	ptr, _ := syscall.UTF16PtrFromString(script)
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 29,
		uintptr(unsafe.Pointer(ptr)),
		handler,
	)
	return com.CheckHR(r, "ExecuteScript")
}

func (w *ICoreWebView2) Reload() error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 31)
	return com.CheckHR(r, "Reload")
}

func (w *ICoreWebView2) PostWebMessageAsJSON(json string) error {
	ptr, _ := syscall.UTF16PtrFromString(json)
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 32,
		uintptr(unsafe.Pointer(ptr)),
	)
	return com.CheckHR(r, "PostWebMessageAsJson")
}

func (w *ICoreWebView2) PostWebMessageAsString(msg string) error {
	ptr, _ := syscall.UTF16PtrFromString(msg)
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 33,
		uintptr(unsafe.Pointer(ptr)),
	)
	return com.CheckHR(r, "PostWebMessageAsString")
}

func (w *ICoreWebView2) AddWebMessageReceivedHandler(handler uintptr) (EventRegistrationToken, error) {
	var token EventRegistrationToken
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 34,
		handler,
		uintptr(unsafe.Pointer(&token)),
	)
	err := com.CheckHR(r, "add_WebMessageReceived")
	fmt.Printf("[ifaces] AddWebMessageReceivedHandler HRESULT=0x%08X token=%v err=%v\n", uint32(r), token, err)
	return token, err
}

func (w *ICoreWebView2) GetDocumentTitle() (string, error) {
	var titlePtr *uint16
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 48,
		uintptr(unsafe.Pointer(&titlePtr)),
	)
	if err := com.CheckHR(r, "get_DocumentTitle"); err != nil {
		return "", err
	}
	return com.UTF16PtrToString(titlePtr), nil
}

func (w *ICoreWebView2) OpenDevToolsWindow() error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 51)
	return com.CheckHR(r, "OpenDevToolsWindow")
}

func (w *ICoreWebView2) AddNavigationCompletedHandler(handler uintptr) (EventRegistrationToken, error) {
	var token EventRegistrationToken
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(w)), 15,
		handler,
		uintptr(unsafe.Pointer(&token)),
	)
	err := com.CheckHR(r, "add_NavigationCompleted")
	fmt.Printf("[ifaces] AddNavigationCompletedHandler HRESULT=0x%08X token=%v err=%v\n", uint32(r), token, err)
	return token, err
}

// ─────────────────────────────────────────────────────────────────
// ICoreWebView2Settings
// vtable indices (after IUnknown: 0,1,2):
//   3  get_IsScriptEnabled
//   4  put_IsScriptEnabled
//   5  get_IsWebMessageEnabled
//   6  put_IsWebMessageEnabled
//   7  get_AreDefaultScriptDialogsEnabled
//   8  put_AreDefaultScriptDialogsEnabled
//   9  get_IsStatusBarEnabled
//  10  put_IsStatusBarEnabled
//  11  get_AreDevToolsEnabled
//  12  put_AreDevToolsEnabled
//  13  get_AreDefaultContextMenusEnabled
//  14  put_AreDefaultContextMenusEnabled
//  15  get_AreHostObjectsAllowed
//  16  put_AreHostObjectsAllowed
//  17  get_IsZoomControlEnabled
//  18  put_IsZoomControlEnabled
//  19  get_IsBuiltInErrorPageEnabled
//  20  put_IsBuiltInErrorPageEnabled
// ─────────────────────────────────────────────────────────────────

type ICoreWebView2Settings struct{ vtable uintptr }

func (s *ICoreWebView2Settings) PutIsScriptEnabled(enabled bool) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(s)), 4, boolToUintptr(enabled))
	return com.CheckHR(r, "put_IsScriptEnabled")
}

func (s *ICoreWebView2Settings) PutIsWebMessageEnabled(enabled bool) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(s)), 6, boolToUintptr(enabled))
	return com.CheckHR(r, "put_IsWebMessageEnabled")
}

func (s *ICoreWebView2Settings) PutIsStatusBarEnabled(enabled bool) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(s)), 10, boolToUintptr(enabled))
	return com.CheckHR(r, "put_IsStatusBarEnabled")
}

func (s *ICoreWebView2Settings) PutAreDevToolsEnabled(enabled bool) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(s)), 12, boolToUintptr(enabled))
	return com.CheckHR(r, "put_AreDevToolsEnabled")
}

func (s *ICoreWebView2Settings) PutAreDefaultContextMenusEnabled(enabled bool) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(s)), 14, boolToUintptr(enabled))
	return com.CheckHR(r, "put_AreDefaultContextMenusEnabled")
}

func (s *ICoreWebView2Settings) PutIsZoomControlEnabled(enabled bool) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(s)), 18, boolToUintptr(enabled))
	return com.CheckHR(r, "put_IsZoomControlEnabled")
}

func (s *ICoreWebView2Settings) PutIsBuiltInErrorPageEnabled(enabled bool) error {
	r, _, _ := com.VTableCall(uintptr(unsafe.Pointer(s)), 20, boolToUintptr(enabled))
	return com.CheckHR(r, "put_IsBuiltInErrorPageEnabled")
}

// ─────────────────────────────────────────────────────────────────
// SUPPORTING TYPES
// ─────────────────────────────────────────────────────────────────

// EventRegistrationToken is returned by add_* event methods.
// Pass to remove_* to unregister.
type EventRegistrationToken struct {
	Value int64
}

// Rect mirrors RECT used by WebView2 controller bounds.
type Rect struct {
	Left, Top, Right, Bottom int32
}

func boolToUintptr(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
