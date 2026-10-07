# WebGo Architecture

This document explains how WebGo works internally and how the major components fit together.

---

# High-level design

WebGo combines four layers:

```text
Application
    │
    ▼
webview package
    │
    ▼
edge package
    │
    ▼
Win32 / COM / WebView2
```

A typical application uses only the public `webview` package.

The lower layers provide native Windows and WebView2 integration.

---

# Repository layout

```text
webview/
├── webview.go
│
├── edge/
│   ├── callbacks.go
│   ├── chromium.go
│   └── ifaces.go
│
├── loader/
│   ├── loader.go
│   └── WebView2Loader.dll
│
└── wAPI/
    ├── com/
    │   └── com.go
    │
    └── w32/
        ├── syscall.go
        ├── types.go
        └── wndproc.go
```

---

# webview.go

This is the application-facing layer.

Responsibilities include:

- native window creation
- WebView lifecycle
- Win32 message handling
- dispatching
- size hints
- navigation queuing
- initialization-script queuing
- JavaScript evaluation
- Go function binding
- JavaScript RPC handling
- navigation policy
- multi-window ownership tracking

Applications should normally depend on this layer rather than the lower-level packages.

---

# edge/chromium.go

`Chromium` owns the WebView2 environment and browser interfaces.

Important native objects include:

```text
ICoreWebView2Environment
        │
        ▼
ICoreWebView2Controller
        │
        ▼
ICoreWebView2
        │
        ▼
ICoreWebView2Settings
```

The WebView2 environment is created asynchronously.

Once the environment exists, WebGo creates the controller for the target HWND.

Once the controller exists, WebGo obtains the actual `ICoreWebView2` interface.

---

# Asynchronous WebView2 initialization

Initialization roughly follows:

```text
NewWithOptionsE
    │
    ▼
webview.init
    │
    ├─ CoInitializeEx
    ├─ register Win32 class
    ├─ create or attach HWND
    ├─ create Chromium wrapper
    └─ Chromium.Embed
            │
            ▼
CreateCoreWebView2EnvironmentWithOptions
            │
            ▼
environment callback
            │
            ▼
CreateCoreWebView2Controller
            │
            ▼
controller callback
            │
            ├─ get ICoreWebView2
            ├─ get settings
            ├─ register events
            ├─ resize browser
            ├─ show controller
            └─ signal OnReady
```

Because this sequence is asynchronous, WebGo queues operations that need WebView2 to exist.

---

# Navigation queuing

Calling:

```go
w.Navigate(url)
```

before WebView2 becomes ready does not immediately call the COM API.

Instead WebGo stores the pending URL.

When `OnReady` fires, the pending URL is consumed and navigation begins.

This allows application code to configure and navigate the WebView immediately after construction.

---

# Initialization script queuing

`Init` behaves similarly.

Initialization scripts are stored in the WebView object before browser readiness.

When WebView2 becomes ready they are registered using:

```text
AddScriptToExecuteOnDocumentCreated
```

This ensures the scripts are available for future documents.

Bindings also use this mechanism so their JavaScript stubs survive navigation.

---

# JavaScript-to-Go bridge

WebGo's binding system is Promise based.

A binding:

```go
w.Bind("add", func(a, b int) int {
    return a + b
})
```

creates a JavaScript function approximately equivalent to:

```js
window.add = function(a, b) {
    return new Promise((resolve, reject) => {
        // store callback
        // send call message to Go
    });
};
```

The call path is:

```text
JavaScript
    │
    │ add(5, 7)
    ▼
Promise stub
    │
    │ chrome.webview.postMessage
    ▼
WebView2 WebMessage event
    │
    ▼
Go bridge parser
    │
    ▼
binding lookup
    │
    ▼
reflection argument conversion
    │
    ▼
Go function
    │
    ▼
JSON result
    │
    ▼
PostWebMessage
    │
    ▼
JavaScript bridge
    │
    ▼
Promise resolve/reject
```

---

# Bridge message format

Internally, binding calls use messages containing fields such as:

```json
{
  "type": "__call__",
  "id": "request-id",
  "name": "add",
  "args": [5, 7]
}
```

Results use:

```json
{
  "type": "__result__",
  "id": "request-id",
  "result": 12
}
```

Errors use:

```json
{
  "type": "__result__",
  "id": "request-id",
  "error": "error message"
}
```

The request ID allows multiple concurrent Promise calls to be matched to their responses.

---

# Reflection conversion

WebGo uses Go reflection and `encoding/json` to convert incoming JavaScript arguments.

For each declared Go parameter:

1. WebGo creates a value of that type.
2. The matching JSON argument is decoded into it.
3. The resulting values are passed into the function with reflection.

This allows strongly typed bindings while keeping the browser protocol JSON based.

---

# Binding panic protection

Bound functions execute behind a recovery boundary.

If a native function panics, WebGo converts that failure into a bridge error instead of intentionally allowing the panic to escape through the COM/WebView event callback.

This is important because panics crossing native callback boundaries can lead to unstable process behavior.

---

# Binding concurrency

Application bindings are not executed inline in the WebView2 message handler.

They run asynchronously so expensive application work does not unnecessarily block the WebView event callback.

This makes the bridge suitable for:

- database queries
- filesystem operations
- network requests
- parsing
- application services

The result is dispatched back to the WebView when ready.

---

# Go-to-JavaScript calls

WebGo provides three primary mechanisms.

## Init

Installs JavaScript for every future document.

## Eval

Runs JavaScript in the current document without returning a result.

## EvalWithResult

Runs JavaScript and invokes a Go completion callback with the JSON-encoded result.

---

# Native WebMessages

`PostMessage` provides lower-level Go-to-JavaScript communication.

This is separate from the binding RPC layer.

It is useful for application-defined event channels.

Example application events:

```text
connection-open
connection-closed
download-progress
database-updated
server-message
```

---

# Navigation policy

WebView2 raises a NavigationStarting event before navigation begins.

WebGo exposes that as:

```go
SetNavigationHandler(func(uri string) bool)
```

Returning false causes WebGo to set WebView2's cancel flag.

This is particularly important because a WebView may have access to privileged native bindings.

---

# COM layer

`wAPI/com` provides pure-Go COM primitives.

Key responsibilities:

- GUID representation
- vtable calls
- HRESULT helpers
- COM callback thunks
- UTF-16 conversion
- COM-allocated memory cleanup

WebView2 COM interfaces are represented as Go structs whose first field points to a COM vtable.

A native method call is performed by:

1. reading the object's vtable pointer
2. selecting the correct method index
3. calling the function pointer with `syscall.SyscallN`

This avoids CGo.

---

# COM callbacks

WebView2 relies heavily on callback interfaces.

Examples include:

- environment creation completed
- controller creation completed
- navigation starting
- navigation completed
- WebMessage received
- ExecuteScript completed

WebGo constructs COM-compatible callback objects in Go.

Each callback provides:

```text
QueryInterface
AddRef
Release
Invoke
```

The callback objects are retained by Go while native WebView2 code may still reference them.

---

# GC anchors

Passing a Go pointer to native code as a raw `uintptr` does not automatically keep the Go object alive.

For that reason `Chromium` stores callback handler objects as fields.

Examples:

```text
envHandler
ctrlHandler
msgHandler
navHandler
navStartingHandler
scriptHandlers
```

These references prevent the Go garbage collector from reclaiming callback structures while WebView2 can still call them.

---

# ExecuteScript callback lifetime

`EvalWithResult` creates an asynchronous COM completion handler.

Those handlers are stored in a map keyed by an internal request ID.

They are removed only after WebView2 invokes the callback or the immediate ExecuteScript call fails.

This prevents the handler from becoming unreachable while native code still owns a pointer to it.

---

# Event registration tokens

WebView2 add-event methods return `EventRegistrationToken` values.

WebGo stores tokens for:

- NavigationStarting
- WebMessageReceived
- NavigationCompleted

During destruction, the corresponding remove-event methods are called before the WebView COM interfaces are released.

---

# Win32 layer

`wAPI/w32` wraps native Windows functionality.

Responsibilities include:

- window class registration
- HWND creation
- message-loop handling
- native messages
- window text
- sizing
- DPI
- dispatch queues
- window-to-WebView context mapping

---

# Window context map

WebGo associates each native HWND with its `webview` object.

Conceptually:

```text
HWND -> *webview
```

The global WndProc can then recover the correct WebView instance for incoming native messages.

---

# Per-window dispatch

Background goroutines cannot safely perform arbitrary UI work directly.

WebGo queues functions per HWND.

```text
goroutine
    │
    ▼
Dispatch(hwnd, fn)
    │
    ├─ queue fn for hwnd
    └─ PostMessage(WM_APP_DISPATCH)
            │
            ▼
        WndProc
            │
            ▼
    DrainDispatch(hwnd)
            │
            ▼
        execute fn
```

The queue is keyed by HWND rather than one global active window, which is required for multiple WebViews.

---

# Resize handling

The WndProc handles `WM_SIZE`.

When the parent window changes size, WebGo asks the controller to update its bounds to match the client rectangle.

This keeps the embedded browser filling the native window.

---

# DPI handling

WebGo uses Windows DPI information when sizing windows.

It also handles `WM_DPICHANGED` and uses the rectangle suggested by Windows when a window moves between displays with different DPI scaling.

---

# Size constraints

`WM_GETMINMAXINFO` is used for:

- fixed-size windows
- minimum-size windows
- maximum-size windows

The values are controlled by the public `SetSize` hints.

---

# Multi-window lifetime

WebGo tracks top-level windows it owns.

An externally supplied HWND is not counted as an owned window.

When an owned window is destroyed, the count is reduced.

The process message loop receives `WM_QUIT` when the final owned WebGo window is destroyed.

This prevents closing one WebView from automatically terminating unrelated WebGo windows.

---

# WebView2 settings

When the browser becomes ready WebGo configures settings including:

- WebMessages enabled
- built-in error page enabled
- zoom controls enabled

The public debug option additionally controls:

- DevTools
- default context menus

The status bar is disabled by the higher-level WebView initialization.

---

# Loader

`loader/loader.go` locates and loads `WebView2Loader.dll`.

It exposes native WebView2 loader functions including:

- CreateCoreWebView2EnvironmentWithOptions
- GetAvailableCoreWebView2BrowserVersionString
- CompareBrowserVersions

The loader also frees strings returned by WebView2 using `CoTaskMemFree`.

---

# User data

The WebView2 environment can receive a custom user-data folder.

That folder controls browser-profile state such as:

- cookies
- cache
- local storage
- IndexedDB
- service-worker data

Applications can use this to isolate browser sessions.

---

# Destruction order

Shutdown is intentionally ordered.

At a high level:

```text
remove WebView2 event handlers
        │
        ▼
release settings
        │
        ▼
close controller
        │
        ▼
release WebView
        │
        ▼
release controller
        │
        ▼
release environment
        │
        ▼
drop Go callback anchors
        │
        ▼
clear pending script handlers
```

The higher-level WebView then removes its dispatch state and uninitializes COM when appropriate.

---

# Why WebGo avoids Electron-style bundling

WebGo uses the WebView2 runtime provided by Microsoft rather than packaging Chromium and Node.js with every application.

The architecture is therefore closer to:

```text
Go application
    +
WebGo
    +
Windows WebView2 Runtime
```

instead of:

```text
application
    +
Node.js
    +
bundled Chromium
    +
desktop framework runtime
```

The tradeoff is that WebGo is Windows-specific and interacts directly with native Win32 and COM concepts.

---

# Intended extension points

The existing architecture can be extended with additional WebView2 interfaces and events.

Potential future areas include:

- virtual-host resource mapping
- resource interception
- download handling
- permission handling
- new-window requests
- custom context menus
- cookies/profile APIs
- process-failure recovery
- custom user agents
- native file dialogs
- system tray integration
- drag and drop
- custom title bars

The `edge/ifaces.go` layer is the natural place to add new WebView2 COM methods.

The `Chromium` layer should own native event registration and lifetime.

The public `webview` layer should expose application-friendly APIs on top.
