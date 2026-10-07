//go:build windows && amd64

package webview

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"syscall"
	"unsafe"

	"webgo_repo-main/webview/edge"
	"webgo_repo-main/webview/loader"
	"webgo_repo-main/webview/wAPI/w32"
)

type Hint int

const (
	HintNone Hint = iota
	HintFixed
	HintMin
	HintMax
)

type WindowOptions struct {
	Title  string
	Width  int32
	Height int32
	IconId int
}
type bridgeMessage struct {
	Type string            `json:"type"`
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Args []json.RawMessage `json:"args"`
}
type WebViewOptions struct {
	Debug          bool
	AutoFocus      bool
	UserDataFolder string
	Window         WindowOptions
	ExistingWindow uintptr
}

type WebView interface {
	Run()
	Terminate()
	Dispatch(fn func())
	Destroy()
	SetTitle(title string)
	SetSize(w, h int32, hint Hint)
	Navigate(url string)
	NavigateToString(html string)
	Init(script string)
	Eval(script string)
	Bind(name string, fn interface{}) error
	Unbind(name string)
	PostMessage(msg string)
	OpenDevTools()
	Version() string
	HWND() uintptr
	EvalDirect(script string)
}

func New(debug bool) WebView {
	syscall.NewLazyDLL("shcore.dll").NewProc("SetProcessDpiAwareness").Call(1)
	return NewWithOptions(WebViewOptions{
		Debug:     debug,
		AutoFocus: true,
		Window:    WindowOptions{Title: "webview", Width: 800, Height: 600},
	})
}

func NewWithOptions(opts WebViewOptions) WebView {
	ver, err := loader.GetInstalledVersion()
	if err != nil {
		panic("webview: cannot check WebView2 runtime: " + err.Error())
	}
	if ver == "" {
		panic("webview: WebView2 Runtime is not installed.\n" +
			"  Download from: https://developer.microsoft.com/en-us/microsoft-edge/webview2/")
	}
	fmt.Print("[Go_custom_webview_wrapper] BongTTY found ")
	slog.Info("WebView2 runtime found", "version", ver)
	wv := &webview{opts: opts, bindings: make(map[string]binding)}
	wv.init()
	return wv
}

const wndClassName = "webview2_window"

type webview struct {
	opts    WebViewOptions
	hwnd    w32.HWND
	browser *edge.Chromium

	hint       Hint
	minW, minH int32
	maxW, maxH int32

	bindingsMu sync.RWMutex
	bindings   map[string]binding

	pendingMu  sync.Mutex
	pendingNav string
	ready      bool

	// Must be stored here — keeps the thunk reachable from GC.
	wndProcCB uintptr
}

type binding struct {
	fn   interface{}
	stub string
}

func (wv *webview) init() {
	hr := w32.CoInitializeEx(0, w32.COINIT_APARTMENTTHREADED)
	if hr < 0 {
		panic(fmt.Sprintf("webview: CoInitializeEx failed: HRESULT 0x%08X", uint32(hr)))
	}

	wv.wndProcCB = syscall.NewCallback(wndProc)
	w32.RegisterWindowClass(wndClassName, wv.wndProcCB)

	var hwnd w32.HWND
	if wv.opts.ExistingWindow != 0 {
		hwnd = w32.HWND(wv.opts.ExistingWindow)
	} else {
		title := wv.opts.Window.Title
		if title == "" {
			title = "webview"
		}
		w, h := wv.opts.Window.Width, wv.opts.Window.Height
		if w == 0 {
			w = 800
		}
		if h == 0 {
			h = 600
		}
		hwnd = w32.CreateMainWindow(wndClassName, title, w, h, w32.GetModuleHandle(""))
		if hwnd == 0 {
			panic("webview: CreateWindowExW failed")
		}
	}
	wv.hwnd = hwnd
	w32.SetWindowContext(hwnd, wv)
	w32.SetDispatchHWND(hwnd)
	dark := 1
	// Call DwmSetWindowAttribute from dwmapi.dll to set immersive dark mode
	syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute").Call(
		uintptr(hwnd),
		uintptr(w32.DWMWA_USE_IMMERSIVE_DARK_MODE),
		uintptr(unsafe.Pointer(&dark)),
		uintptr(unsafe.Sizeof(dark)),
	)
	wv.browser = edge.NewChromium(hwnd)

	// ── Set ALL callbacks BEFORE Embed ────────────────────────────
	// These must be set before Embed() because the controller is
	// created asynchronously inside Embed and registers handlers
	// immediately when the controller callback fires.

	wv.browser.OnMessage(func(rawMsg string) {
		wv.handleJSMessage(rawMsg)
	})

	wv.browser.OnNavigationCompleted(func(success bool) {
		if !success {
			return
		}
		// Post script injection back to message loop,
		// not directly from inside the navigation callback
		w32.Dispatch(func() {
			fmt.Println("[Go] Injecting bridge script via dispatch")
			wv.browser.Eval(jsBridgeScript)

			wv.bindingsMu.RLock()
			count := len(wv.bindings)
			for _, b := range wv.bindings {
				wv.browser.Eval(b.stub)
			}
			wv.bindingsMu.RUnlock()
			fmt.Println("[Go] Injected bridge +", count, "stubs")
		})
	})

	wv.browser.OnReady(func() {
		fmt.Println("[Go] OnReady fired")
		wv.browser.SetDevToolsEnabled(wv.opts.Debug)
		wv.browser.SetContextMenusEnabled(wv.opts.Debug)
		wv.browser.SetStatusBarEnabled(false)

		if wv.opts.AutoFocus {
			wv.browser.Focus()
		}

		wv.pendingMu.Lock()
		wv.ready = true
		nav := wv.pendingNav
		wv.pendingNav = ""
		wv.pendingMu.Unlock()

		fmt.Println("[Go] OnReady: pendingNav =", nav)
		if nav != "" {
			wv.browser.Navigate(nav)
		}
	})

	// ── NOW start the async init ──────────────────────────────────
	fmt.Println("[Go] Calling Embed")
	if err := wv.browser.Embed(wv.opts.UserDataFolder); err != nil {
		panic("webview: failed to embed WebView2: " + err.Error())
	}
	fmt.Println("[Go] Embed returned (async, waiting for callbacks)")

	w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL)
	w32.UpdateWindow(hwnd)
}

// wndProc — ALL params uintptr, see comment in init().
// wndProc — registered as a free function so it's safe to call
// during window creation before wv is stored in the context map.
func wndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	// Look up the webview for this HWND. May be nil during creation
	// messages (WM_GETMINMAXINFO etc.) sent before SetWindowContext runs.
	wv, _ := w32.GetWindowContext(w32.HWND(hwnd)).(*webview)
	if wv == nil {
		return uintptr(w32.DefWindowProc(w32.HWND(hwnd), uint32(msg), w32.WPARAM(wParam), w32.LPARAM(lParam)))
	}

	switch uint32(msg) {
	case w32.WM_SIZE:
		if wv != nil && wv.browser != nil && uint32(wParam) != w32.SIZE_MINIMIZED {
			wv.browser.Resize()
		}
		return 0

	case w32.WM_GETMINMAXINFO:
		if wv == nil {
			break // let DefWindowProc handle it during creation
		}
		mmi := (*w32.MinMaxInfo)(unsafe.Pointer(lParam))
		switch wv.hint {
		case HintFixed:
			mmi.MinTrackSize.X = wv.minW
			mmi.MinTrackSize.Y = wv.minH
			mmi.MaxTrackSize.X = wv.minW
			mmi.MaxTrackSize.Y = wv.minH
		case HintMin:
			mmi.MinTrackSize.X = wv.minW
			mmi.MinTrackSize.Y = wv.minH
		case HintMax:
			mmi.MaxTrackSize.X = wv.maxW
			mmi.MaxTrackSize.Y = wv.maxH
		}
		return 0

	case w32.WM_APP_DISPATCH:
		w32.DrainDispatch()
		return 0

	case w32.WM_CLOSE:
		w32.DestroyWindow(w32.HWND(hwnd))
		return 0

	case w32.WM_DESTROY:
		w32.DeleteWindowContext(w32.HWND(hwnd))
		w32.PostQuitMessage(0)
		return 0
	case 0x02E0: // WM_DPICHANGED
		// lParam points to a RECT with the suggested new window size
		rect := (*w32.Rect)(unsafe.Pointer(lParam))
		procSetWindowPos.Call(
			uintptr(hwnd),
			0,
			uintptr(rect.Left),
			uintptr(rect.Top),
			uintptr(rect.Right-rect.Left),
			uintptr(rect.Bottom-rect.Top),
			0x0004|0x0010, // SWP_NOZORDER | SWP_NOACTIVATE
		)
		if wv != nil && wv.browser != nil {
			wv.browser.Resize()
		}
		return 0
	}

	return uintptr(w32.DefWindowProc(w32.HWND(hwnd), uint32(msg), w32.WPARAM(wParam), w32.LPARAM(lParam)))
}

func (wv *webview) Run()               { w32.RunMessageLoop(wv.hwnd) }
func (wv *webview) Terminate()         { w32.PostMessage(wv.hwnd, w32.WM_CLOSE, 0, 0) }
func (wv *webview) Dispatch(fn func()) { w32.Dispatch(fn) }
func (wv *webview) SetTitle(t string)  { w32.SetWindowText(wv.hwnd, t) }
func (wv *webview) Version() string    { return wv.browser.BrowserVersion() }
func (wv *webview) HWND() uintptr      { return uintptr(wv.hwnd) }

func (wv *webview) Destroy() {
	if wv.browser != nil {
		wv.browser.Destroy()
		wv.browser = nil
	}
	w32.CoUninitialize()
}

var procSetWindowPos = syscall.NewLazyDLL("user32.dll").NewProc("SetWindowPos")

func (wv *webview) SetSize(width, height int32, hint Hint) {
	// Get DPI for this window
	user32 := syscall.NewLazyDLL("user32.dll")
	getDpiForWindow := user32.NewProc("GetDpiForWindow")
	dpi, _, _ := getDpiForWindow.Call(uintptr(wv.hwnd))
	if dpi == 0 {
		dpi = 96
	}

	// Scale logical → physical
	physW := (width * int32(dpi)) / 96
	physH := (height * int32(dpi)) / 96

	procSetWindowPos.Call(
		uintptr(wv.hwnd), 0,
		0, 0,
		uintptr(physW), uintptr(physH),
		0x0002|0x0004|0x0010,
	)
	if wv.browser != nil {
		wv.browser.Resize()
	}
}

func (wv *webview) Navigate(url string) {
	wv.pendingMu.Lock()

	if !wv.ready {
		wv.pendingNav = url
		wv.pendingMu.Unlock()
		return
	}

	wv.pendingMu.Unlock()

	wv.Dispatch(func() {
		if wv.browser != nil {
			wv.browser.Navigate(url)
		}
	})
}

func (wv *webview) navigateDirect(url string) {
	wv.pendingMu.Lock()

	if !wv.ready {
		wv.pendingNav = url
		wv.pendingMu.Unlock()
		return
	}

	wv.pendingMu.Unlock()

	wv.Dispatch(func() {
		if wv.browser != nil {
			wv.browser.Navigate(url)
		}
	})
}

func (wv *webview) NavigateToString(html string) {
	wv.Dispatch(func() { wv.browser.NavigateToString(html) })
}

// Public versions dispatch for cross-thread safety
func (wv *webview) Init(script string) { wv.Dispatch(func() { wv.initDirect(script) }) }
func (wv *webview) Eval(script string) { wv.Dispatch(func() { wv.evalDirect(script) }) }

func (wv *webview) PostMessage(msg string) {
	wv.Dispatch(func() { wv.browser.PostMessage(msg) })
}

func (wv *webview) OpenDevTools() {
	wv.Dispatch(func() { wv.browser.OpenDevTools() })
}

func (wv *webview) Bind(name string, fn interface{}) error {
	if name == "" {
		return fmt.Errorf("webview: Bind: name cannot be empty")
	}

	if fn == nil {
		return fmt.Errorf("webview: Bind: %q function is nil", name)
	}

	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return fmt.Errorf("webview: Bind: %q is not a function", name)
	}

	// JSON encoding produces a valid JavaScript string literal.
	nameJSON, err := json.Marshal(name)
	if err != nil {
		return fmt.Errorf("webview: Bind: invalid name %q: %w", name, err)
	}

	jsName := string(nameJSON)

	stub := fmt.Sprintf(`
(function () {
	const name = %s;

	window[name] = function () {
		const id =
			Date.now().toString(36) + "-" +
			Math.random().toString(36).slice(2);

		const args = Array.prototype.slice.call(arguments);

		return new Promise(function (resolve, reject) {
			window.__wv2_cbs = window.__wv2_cbs || {};

			window.__wv2_cbs[id] = {
				resolve: resolve,
				reject: reject
			};

			window.chrome.webview.postMessage(JSON.stringify({
				type: "__call__",
				id: id,
				name: name,
				args: args
			}));
		});
	};
})();
`, jsName)

	wv.bindingsMu.Lock()
	wv.bindings[name] = binding{
		fn:   fn,
		stub: stub,
	}
	wv.bindingsMu.Unlock()

	if wv.browser != nil {
		wv.browser.Init(stub)

		wv.pendingMu.Lock()
		ready := wv.ready
		wv.pendingMu.Unlock()

		if ready {
			wv.Eval(stub)
		}
	}

	return nil
}

func (wv *webview) Unbind(name string) {
	wv.bindingsMu.Lock()
	delete(wv.bindings, name)
	wv.bindingsMu.Unlock()

	nameJSON, err := json.Marshal(name)
	if err != nil {
		return
	}

	wv.Eval(fmt.Sprintf(
		`delete window[%s];`,
		string(nameJSON),
	))
}

func (wv *webview) handleJSMessage(raw string) {
	var msg bridgeMessage

	// WebView2 may give us a JSON encoded string containing our JSON.
	var encoded string

	if err := json.Unmarshal([]byte(raw), &encoded); err == nil {
		if err := json.Unmarshal([]byte(encoded), &msg); err != nil {
			slog.Error(
				"invalid WebView bridge message",
				"err", err,
			)
			return
		}
	} else {
		// Or it may already be the JSON object.
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			slog.Error(
				"invalid WebView bridge message",
				"err", err,
			)
			return
		}
	}

	if msg.Type == "" {
		return
	}

	wv.processJSMessage(msg)
}
func (wv *webview) sendBindingError(id string, err error) {
	response := struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Error string `json:"error"`
	}{
		Type:  "__result__",
		ID:    id,
		Error: err.Error(),
	}

	data, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		return
	}

	wv.Dispatch(func() {
		if wv.browser != nil {
			wv.browser.PostMessage(string(data))
		}
	})
}

// processJSMessage handles the parsed message and calls the bound Go function.
func (wv *webview) processJSMessage(msg bridgeMessage) {
	if msg.Type != "__call__" {
		return
	}

	wv.bindingsMu.RLock()
	b, ok := wv.bindings[msg.Name]
	wv.bindingsMu.RUnlock()

	if !ok {
		wv.sendBindingError(
			msg.ID,
			fmt.Errorf("binding %q does not exist", msg.Name),
		)
		return
	}

	// Never execute arbitrary bound functions inside the WebView2 UI callback.
	go func() {
		result, err := callBinding(b.fn, msg.Args)

		var response struct {
			Type   string      `json:"type"`
			ID     string      `json:"id"`
			Result interface{} `json:"result,omitempty"`
			Error  string      `json:"error,omitempty"`
		}

		response.Type = "__result__"
		response.ID = msg.ID

		if err != nil {
			response.Error = err.Error()
		} else {
			response.Result = result
		}

		data, marshalErr := json.Marshal(response)
		if marshalErr != nil {
			wv.sendBindingError(msg.ID, marshalErr)
			return
		}

		wv.Dispatch(func() {
			if wv.browser != nil {
				wv.browser.PostMessage(string(data))
			}
		})
	}()
}

func callBinding(fn interface{}, rawArgs []json.RawMessage) (interface{}, error) {
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()
	if fnType.NumIn() != len(rawArgs) {
		return nil, fmt.Errorf("expected %d args, got %d", fnType.NumIn(), len(rawArgs))
	}
	args := make([]reflect.Value, fnType.NumIn())
	for i := range args {
		p := reflect.New(fnType.In(i))
		if err := json.Unmarshal(rawArgs[i], p.Interface()); err != nil {
			return nil, fmt.Errorf("arg %d: %w", i, err)
		}
		args[i] = p.Elem()
	}
	results := fnVal.Call(args)
	if len(results) == 0 {
		return nil, nil
	}
	last := results[len(results)-1]
	if last.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !last.IsNil() {
			return nil, last.Interface().(error)
		}
		results = results[:len(results)-1]
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0].Interface(), nil
}

// jsBridgeScript is injected on every page load.
// It listens for __result__ messages posted by Go and resolves/rejects
// the Promise that the Bind() stub created.
const jsBridgeScript = `(function() {
  window.__wv2_cbs = {};
  window.chrome.webview.addEventListener('message', function(e) {
    try {
      var msg = JSON.parse(e.data);
      if (!msg || msg.type !== '__result__') return;
      var cb = window.__wv2_cbs[msg.id];
      if (!cb) return;
      delete window.__wv2_cbs[msg.id];
      if (msg.error !== undefined) cb.reject(new Error(msg.error));
      else cb.resolve(msg.result !== undefined ? msg.result : null);
    } catch(_) {}
  });
})();`

// Add direct versions for use on the UI thread
func (wv *webview) initDirect(script string) { wv.browser.Init(script) }
func (wv *webview) evalDirect(script string) { wv.browser.Eval(script) }

// Implementation:
func (wv *webview) EvalDirect(script string) {
	if wv.browser != nil {
		wv.browser.Eval(script)
	}
}
