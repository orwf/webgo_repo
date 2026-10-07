//go:build windows && amd64

package edge

import (
	"fmt"
	"log/slog"
	"sync"
	"syscall"
	"unsafe"

	"webgo_repo-main/webview/loader"
	"webgo_repo-main/webview/wAPI/com"
	"webgo_repo-main/webview/wAPI/w32"
)

// Chromium manages the Edge/WebView2 subprocess embedded in a Win32 window.
type Chromium struct {
	hwnd               w32.HWND
	environment        *ICoreWebView2Environment
	controller         *ICoreWebView2Controller
	webview            *ICoreWebView2
	settings           *ICoreWebView2Settings
	navStartingHandler *navigationStartingHandler

	onMessageReceived     func(string)
	onNavigationCompleted func(bool)
	onReady               func()

	// *** GC ANCHOR — critical ***
	//
	// envHandler and ctrlHandler are passed as raw uintptr pointers to the
	// WebView2 DLL which calls back into them asynchronously. Once Embed()
	// returns, there is no Go pointer to these locals, so the GC is free to
	// collect them before the callbacks fire. Storing them on the struct
	// keeps them alive for the lifetime of the Chromium object.
	envHandler  *environmentCompletedHandler
	ctrlHandler *controllerCompletedHandler
	// Event handlers registered on the webview also need to stay alive.
	msgHandler *webMessageReceivedHandler
	navHandler *navigationCompletedHandler

	scriptMu       sync.Mutex
	scriptNextID   uint64
	scriptHandlers map[uint64]*executeScriptCompletedHandler
}

func NewChromium(hwnd w32.HWND) *Chromium {
	return &Chromium{
		hwnd:           hwnd,
		scriptHandlers: make(map[uint64]*executeScriptCompletedHandler),
	}
}

// Embed initialises the WebView2 environment and embeds the browser into the
// parent hwnd. The actual work is async — callbacks fire during the Win32
// message loop. When the controller is ready, OnReady() is called.
func (c *Chromium) Embed(userDataFolder string) error {
	var udFolder *uint16
	if userDataFolder != "" {
		udFolder, _ = syscall.UTF16PtrFromString(userDataFolder)
	}

	// Create and STORE the env handler on the struct to prevent GC collection
	// before the async callback fires.
	c.envHandler = NewEnvironmentCompletedHandler(func(result uintptr, env *ICoreWebView2Environment) {
		if int32(result) < 0 {
			slog.Error("WebView2 environment creation failed", "hresult", fmt.Sprintf("0x%08X", uint32(result)))
			return
		}
		c.environment = env
		env.AddRef()

		// Same for the controller handler.
		c.ctrlHandler = NewControllerCompletedHandler(func(result uintptr, ctrl *ICoreWebView2Controller) {
			if int32(result) < 0 {
				slog.Error("WebView2 controller creation failed",
					"hresult", fmt.Sprintf("0x%08X", uint32(result)))
				return
			}
			c.controller = ctrl
			ctrl.AddRef()

			wv, err := ctrl.GetICoreWebView2()
			if err != nil {
				slog.Error("GetICoreWebView2 failed", "err", err)
				return
			}
			c.webview = wv
			wv.AddRef()

			// <-- ADD HERE: Register NavigationStarting handler
			c.navStartingHandler = NewNavigationStartingHandler(func(args *ICoreWebView2NavigationStartingEventArgs) {
				uri, err := args.GetUri()
				if err != nil {
					fmt.Println("[Chromium] NavigationStarting: error getting URI:", err)
					return
				}
				fmt.Println("[Chromium] NavigationStarting to:", uri)
			})
			tok, err := wv.AddNavigationStartingHandler(c.navStartingHandler.AsPtr())
			fmt.Println("[Chromium] AddNavigationStartingHandler token:", tok, "err:", err)

			if settings, err := wv.GetSettings(); err == nil {
				c.settings = settings
			} else {
				slog.Error("GetSettings failed", "err", err)
			}
			c.applyDefaultSettings()
			c.Resize()
			ctrl.PutIsVisible(true)

			// ── ALWAYS register message handler ──────────────────────────
			// Check the callback at invocation time, not registration time.
			c.msgHandler = NewWebMessageReceivedHandler(func(msg string) {
				if c.onMessageReceived != nil {
					c.onMessageReceived(msg)
				}
			})
			tok, err = wv.AddWebMessageReceivedHandler(c.msgHandler.AsPtr())
			fmt.Println("[Chromium] AddWebMessageReceivedHandler token:", tok, "err:", err)

			// ── ALWAYS register navigation completed handler ──────────────
			c.navHandler = NewNavigationCompletedHandler(func(sender uintptr, args *ICoreWebView2NavigationCompletedEventArgs) {
				// Get the success bool (vtable index 3 = get_IsSuccess)
				var isSuccess int32
				com.VTableCall(uintptr(unsafe.Pointer(args)), 3, uintptr(unsafe.Pointer(&isSuccess)))
				success := isSuccess != 0
				fmt.Println("[Chromium] NavigationCompleted success:", success)
				if c.onNavigationCompleted != nil {
					c.onNavigationCompleted(success)
				}
			})
			tok2, err2 := wv.AddNavigationCompletedHandler(c.navHandler.AsPtr())
			fmt.Println("[Chromium] AddNavigationCompletedHandler token:", tok2, "err:", err2)
			fmt.Print("[Go_custom_chromium_wrapper] Chromium found ")
			slog.Info("WebView2 ready")
			fmt.Println("[Chromium] Calling onReady")
			if c.onReady != nil {
				c.onReady()
			}
			fmt.Println("[Chromium] onReady returned")
		})

		env.CreateCoreWebView2Controller(uintptr(c.hwnd), c.ctrlHandler.AsPtr())
	})

	_, err := loader.CreateEnvironmentWithOptions(nil, udFolder, 0, c.envHandler.AsPtr())
	return err
}

func (c *Chromium) Navigate(url string) {
	if c.webview == nil {
		return
	}
	if err := c.webview.Navigate(url); err != nil {
		slog.Error("Navigate failed", "url", url, "err", err)
	}
}

func (c *Chromium) NavigateToString(html string) {
	if c.webview == nil {
		return
	}
	if err := c.webview.NavigateToString(html); err != nil {
		slog.Error("NavigateToString failed", "err", err)
	}
}

func (c *Chromium) Init(script string) {
	if c.webview == nil {
		return
	}

	c.webview.AddScriptToExecuteOnDocumentCreated(script, 0)
}

func (c *Chromium) Eval(script string) {
	if c.webview == nil {
		return
	}

	if err := c.webview.ExecuteScript(script, 0); err != nil {
		slog.Error(
			"ExecuteScript failed",
			"err", err,
		)
	}
}

func (c *Chromium) EvalWithResult(script string, fn func(string)) {
	if c.webview == nil {
		return
	}

	c.scriptMu.Lock()
	c.scriptNextID++
	id := c.scriptNextID
	c.scriptMu.Unlock()

	var handler *executeScriptCompletedHandler

	handler = NewExecuteScriptCompletedHandler(func(result string) {
		if fn != nil {
			fn(result)
		}

		c.scriptMu.Lock()
		delete(c.scriptHandlers, id)
		c.scriptMu.Unlock()
	})

	c.scriptMu.Lock()
	c.scriptHandlers[id] = handler
	c.scriptMu.Unlock()

	if err := c.webview.ExecuteScript(script, handler.AsPtr()); err != nil {
		c.scriptMu.Lock()
		delete(c.scriptHandlers, id)
		c.scriptMu.Unlock()

		slog.Error(
			"ExecuteScript failed",
			"err", err,
		)
	}
}

func (c *Chromium) PostMessage(msg string) {
	if c.webview == nil {
		return
	}
	c.webview.PostWebMessageAsString(msg)
}

func (c *Chromium) PostJSON(json string) {
	if c.webview == nil {
		return
	}
	c.webview.PostWebMessageAsJSON(json)
}

func (c *Chromium) OnMessage(fn func(string)) {
	c.onMessageReceived = fn
}

func (c *Chromium) OnNavigationCompleted(fn func(bool)) {
	c.onNavigationCompleted = fn
}

func (c *Chromium) OnReady(fn func()) {
	c.onReady = fn
}

func (c *Chromium) Resize() {
	if c.controller == nil {
		return
	}
	var rect w32.Rect
	w32.GetClientRect(c.hwnd, &rect)
	c.controller.PutBounds(Rect{
		Left:   rect.Left,
		Top:    rect.Top,
		Right:  rect.Right,
		Bottom: rect.Bottom,
	})
	c.controller.NotifyParentWindowPositionChanged()
}

func (c *Chromium) Focus() {
	if c.controller != nil {
		c.controller.MoveFocus(0) // COREWEBVIEW2_MOVE_FOCUS_REASON_PROGRAMMATIC
	}
}

func (c *Chromium) OpenDevTools() {
	if c.webview != nil {
		c.webview.OpenDevToolsWindow()
	}
}

func (c *Chromium) applyDefaultSettings() {
	if c.settings == nil {
		return
	}
	c.settings.PutIsWebMessageEnabled(true)
	c.settings.PutIsBuiltInErrorPageEnabled(true)
	c.settings.PutIsZoomControlEnabled(true)
}

func (c *Chromium) SetDevToolsEnabled(enabled bool) {
	if c.settings != nil {
		c.settings.PutAreDevToolsEnabled(enabled)
	}
}

func (c *Chromium) SetContextMenusEnabled(enabled bool) {
	if c.settings != nil {
		c.settings.PutAreDefaultContextMenusEnabled(enabled)
	}
}

func (c *Chromium) SetStatusBarEnabled(enabled bool) {
	if c.settings != nil {
		c.settings.PutIsStatusBarEnabled(enabled)
	}
}

func (c *Chromium) Destroy() {
	if c.controller != nil {
		c.controller.Close()
		c.controller.Release()
		c.controller = nil
	}
	if c.webview != nil {
		c.webview.Release()
		c.webview = nil
	}
	if c.environment != nil {
		c.environment.Release()
		c.environment = nil
	}
	// Nil out handler references so GC can reclaim them after destroy.
	c.envHandler = nil
	c.ctrlHandler = nil
	c.msgHandler = nil
	c.navHandler = nil
	c.navStartingHandler = nil
	c.scriptMu.Lock()
	c.scriptHandlers = make(map[uint64]*executeScriptCompletedHandler)
	c.scriptMu.Unlock()
}

func (c *Chromium) BrowserVersion() string {
	if c.environment == nil {
		ver, _ := loader.GetInstalledVersion()
		return ver
	}
	ver, _ := c.environment.GetBrowserVersionString()
	return ver
}

func unsafePtr(p interface{}) uintptr {
	type iface struct{ _, data unsafe.Pointer }
	return uintptr((*iface)(unsafe.Pointer(&p)).data)
}
