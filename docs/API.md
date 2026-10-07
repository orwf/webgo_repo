# WebGo API Reference

This document describes the public WebGo API as implemented in the current `webview` package.

> Platform: Windows amd64  
> Renderer: Microsoft Edge WebView2  
> Native layer: Win32 + COM  
> Frontend: HTML, CSS, JavaScript

---

## Package overview

WebGo embeds Microsoft Edge WebView2 into a native Win32 window and provides a Go-to-JavaScript bridge.

The public API is centered around the `WebView` interface:

```go
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
    EvalWithResult(script string, fn func(string))
    EvalDirect(script string)

    Bind(name string, fn interface{}) error
    Unbind(name string)

    PostMessage(msg string)

    SetNavigationHandler(fn func(uri string) bool)

    OpenDevTools()

    Version() string
    HWND() uintptr
}
```

---

# Creating a WebView

## New

```go
func New(debug bool) WebView
```

Creates a WebView using default settings.

Defaults:

- title: `webview`
- width: `800`
- height: `600`
- autofocus: enabled
- debug mode: controlled by the argument

Example:

```go
w := webview.New(true)
defer w.Destroy()

w.Navigate("https://example.com")
w.Run()
```

When debug mode is enabled, WebGo enables WebView2 developer tools and default context menus.

---

## NewWithOptions

```go
func NewWithOptions(opts WebViewOptions) WebView
```

Creates a WebView with custom settings.

This constructor panics if creation fails. For applications that want normal error handling, prefer `NewWithOptionsE`.

Example:

```go
w := webview.NewWithOptions(webview.WebViewOptions{
    Debug: true,
    AutoFocus: true,
    UserDataFolder: "./webview-profile",
    Window: webview.WindowOptions{
        Title: "My WebGo App",
        Width: 1200,
        Height: 800,
    },
})
defer w.Destroy()
```

---

## NewWithOptionsE

```go
func NewWithOptionsE(opts WebViewOptions) (WebView, error)
```

Creates a WebView and returns an error instead of panicking.

Recommended for production applications.

Example:

```go
w, err := webview.NewWithOptionsE(webview.WebViewOptions{
    Debug: false,
    AutoFocus: true,
    Window: webview.WindowOptions{
        Title: "Production App",
        Width: 1280,
        Height: 720,
    },
})
if err != nil {
    log.Fatal(err)
}
defer w.Destroy()
```

---

# WebViewOptions

```go
type WebViewOptions struct {
    Debug          bool
    AutoFocus      bool
    UserDataFolder string
    Window         WindowOptions
    ExistingWindow uintptr
}
```

## Debug

Controls development features.

When enabled WebGo enables:

- WebView2 DevTools
- default browser context menus

Use `false` for production applications unless you explicitly want those features available.

---

## AutoFocus

If true, WebGo moves keyboard focus into the WebView when the WebView2 controller becomes ready.

---

## UserDataFolder

Overrides the WebView2 user-data directory.

The user-data folder may contain:

- cookies
- cache
- local storage
- IndexedDB
- service-worker data
- browser profile state

Example:

```go
UserDataFolder: "./profile"
```

Applications that need isolated browser profiles can assign a separate folder per application or account.

---

## ExistingWindow

Embeds WebView2 into an existing native HWND instead of creating a new top-level window.

Example:

```go
opts.ExistingWindow = existingHWND
```

When an existing HWND is used, WebGo does not treat that HWND as one of its owned top-level windows.

---

# WindowOptions

```go
type WindowOptions struct {
    Title  string
    Width  int32
    Height int32
    IconId int
}
```

Current window options include the initial title and dimensions.

---

# Window lifecycle

## Run

```go
w.Run()
```

Starts the native Win32 message loop.

For a normal desktop program this is usually the final call in `main`.

Example:

```go
func main() {
    w := webview.New(false)
    defer w.Destroy()

    w.Navigate("https://example.com")
    w.Run()
}
```

---

## Terminate

```go
w.Terminate()
```

Requests that the WebView's native window close.

This posts a window-close message to the WebView HWND.

---

## Destroy

```go
w.Destroy()
```

Releases WebView2 and native resources owned by the WebView instance.

Cleanup includes:

- WebView2 event handlers
- settings interface
- WebView interface
- controller
- environment
- pending script callback references
- per-window dispatch state
- COM apartment cleanup when WebGo initialized it

Call `Destroy` exactly once for each successfully created WebView.

---

# Thread dispatching

## Dispatch

```go
w.Dispatch(func() {
    // UI-thread work
})
```

Schedules a function to execute on the WebView's native UI thread.

WebView2 and Win32 operations frequently have thread-affinity requirements, so `Dispatch` is important when work originates from goroutines.

Example:

```go
go func() {
    result := doBackgroundWork()

    w.Dispatch(func() {
        w.SetTitle(result)
    })
}()
```

Each WebView maintains dispatching against its own HWND, allowing independent multi-window operation.

---

# Window title

## SetTitle

```go
w.SetTitle("My Application")
```

Changes the native Win32 window title.

---

# Window sizing

## SetSize

```go
w.SetSize(width, height, hint)
```

Resizes the native window and optionally applies sizing constraints.

Available hints:

```go
webview.HintNone
webview.HintFixed
webview.HintMin
webview.HintMax
```

### HintNone

Sets the current size without imposing a permanent minimum or maximum.

```go
w.SetSize(1200, 800, webview.HintNone)
```

### HintFixed

Locks the window to one size.

```go
w.SetSize(900, 600, webview.HintFixed)
```

### HintMin

Sets the minimum resize dimensions.

```go
w.SetSize(640, 480, webview.HintMin)
```

### HintMax

Sets the maximum resize dimensions.

```go
w.SetSize(1920, 1080, webview.HintMax)
```

WebGo applies window DPI scaling before passing physical dimensions to Win32.

---

# Navigation

## Navigate

```go
w.Navigate("https://example.com")
```

Navigates WebView2 to a URL.

Navigation requests made before WebView2 is ready are queued and performed after initialization completes.

---

## NavigateToString

```go
w.NavigateToString(html)
```

Loads an HTML document directly from a Go string.

Example:

```go
w.NavigateToString(`
<!doctype html>
<html>
<body>
    <h1>Hello from WebGo</h1>
</body>
</html>
`)
```

This is useful for:

- small embedded applications
- generated interfaces
- diagnostics
- splash screens
- self-contained tools

For larger applications, serving or embedding separate HTML/CSS/JS resources is usually easier to maintain.

---

# Navigation filtering

## SetNavigationHandler

```go
w.SetNavigationHandler(func(uri string) bool {
    return true
})
```

Registers a navigation policy callback.

Return:

- `true` to allow navigation
- `false` to cancel navigation

Example: only allow a trusted application origin.

```go
w.SetNavigationHandler(func(uri string) bool {
    if strings.HasPrefix(uri, "https://app.example/") {
        return true
    }

    if uri == "about:blank" {
        return true
    }

    return false
})
```

This is one of the most important security features when native Go bindings are exposed to JavaScript.

A remote page that can access privileged bindings could otherwise potentially call native application functions.

---

# Initialization JavaScript

## Init

```go
w.Init(script)
```

Registers JavaScript to execute whenever a new document is created.

Initialization scripts are retained and registered with WebView2 when the browser becomes ready.

Example:

```go
w.Init(`
window.app = {
    name: "WebGo",
    version: 1
};
`)
```

Common uses:

- defining frontend APIs
- setting configuration values
- exposing environment information
- installing event listeners
- polyfills
- preparing global application state

Unlike `Eval`, initialization scripts apply to future document loads.

---

# Executing JavaScript

## Eval

```go
w.Eval(`
document.body.classList.add("connected");
`)
```

Executes JavaScript in the current document.

The call is dispatched onto the WebView UI thread.

Use this for commands where no JavaScript result is required.

---

## EvalDirect

```go
w.EvalDirect(script)
```

Executes JavaScript immediately through the underlying browser object without using the normal public dispatch path.

This is primarily useful when you already know you are running on the WebView UI thread.

For ordinary application code, prefer `Eval`.

---

## EvalWithResult

```go
w.EvalWithResult(`
document.title
`, func(result string) {
    fmt.Println(result)
})
```

Executes JavaScript and provides WebView2's JSON-encoded result to a Go callback.

Example:

```go
w.EvalWithResult(`
({
    title: document.title,
    url: location.href
})
`, func(result string) {
    fmt.Println("JavaScript returned:", result)
})
```

WebView2 returns JavaScript execution values as JSON text.

For example:

JavaScript:

```js
document.title
```

may return:

```json
"My Page"
```

The completion handler is retained internally until WebView2 finishes the asynchronous execution.

---

# Go functions exposed to JavaScript

## Bind

```go
err := w.Bind("functionName", goFunction)
```

Exposes a Go function as a Promise-based JavaScript function.

Example:

```go
w.Bind("greet", func(name string) string {
    return "Hello " + name
})
```

JavaScript:

```js
const value = await greet("Alex");
console.log(value);
```

Result:

```text
Hello Alex
```

WebGo serializes JavaScript arguments through JSON and converts them into the Go function's declared parameter types using reflection.

---

## Supported argument types

Any type that `encoding/json` can decode into the declared Go parameter can be used.

Examples include:

```go
string
bool
int
int64
float64
[]string
[]int
map[string]interface{}
struct
[]MyStruct
```

Example:

```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

w.Bind("saveUser", func(user User) error {
    fmt.Printf("%+v\n", user)
    return nil
})
```

JavaScript:

```js
await saveUser({
    name: "Alice",
    age: 32
});
```

---

# Binding return values

A bound function can return no value:

```go
w.Bind("logMessage", func(message string) {
    log.Println(message)
})
```

A single value:

```go
w.Bind("add", func(a, b int) int {
    return a + b
})
```

JavaScript:

```js
const result = await add(5, 7);
```

A value and an error:

```go
w.Bind("divide", func(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }

    return a / b, nil
})
```

JavaScript:

```js
try {
    const result = await divide(10, 0);
} catch (err) {
    console.error(err.message);
}
```

Multiple non-error return values are returned to JavaScript as an array.

Example:

```go
w.Bind("stats", func() (int, string, bool) {
    return 42, "ready", true
})
```

JavaScript:

```js
const [count, state, ready] = await stats();
```

If a bound Go function panics, WebGo recovers the panic and turns it into an error response instead of allowing the panic to cross the WebView callback boundary.

---

# Binding execution model

Bound Go functions are executed outside the WebView2 message callback.

This prevents long-running native work from unnecessarily blocking the renderer/UI callback.

Examples of appropriate binding work:

- SQL queries
- TCP requests
- HTTP requests
- filesystem access
- parsing
- compression
- application business logic

Example:

```go
w.Bind("searchDatabase", func(query string) ([]Result, error) {
    return database.Search(query)
})
```

JavaScript:

```js
const results = await searchDatabase("example");
```

---

## Unbind

```go
w.Unbind("greet")
```

Removes a previously registered Go binding and deletes the corresponding JavaScript function from the current document.

---

# Raw WebView messages

## PostMessage

```go
w.PostMessage("hello")
```

Posts a string message from Go to WebView2 using the browser's WebMessage channel.

This is separate from the Promise-based `Bind` RPC bridge and can be used for application-specific message protocols.

JavaScript can receive messages with:

```js
window.chrome.webview.addEventListener("message", event => {
    console.log(event.data);
});
```

---

# Developer tools

## OpenDevTools

```go
w.OpenDevTools()
```

Opens the Chromium/Edge developer tools window.

Useful for:

- JavaScript debugging
- network inspection
- DOM inspection
- console output
- storage inspection
- performance profiling

---

# Version information

## Version

```go
version := w.Version()
```

Returns the WebView2 browser/runtime version.

Example:

```go
fmt.Println("WebView2:", w.Version())
```

---

# Native HWND access

## HWND

```go
hwnd := w.HWND()
```

Returns the underlying native Windows HWND.

This allows advanced applications to use additional Win32 APIs against the WebGo window.

Examples:

- custom title bars
- transparency
- native menus
- window positioning
- DWM attributes
- taskbar features
- drag-and-drop
- custom input handling

Because this exposes the native handle directly, application code is responsible for using it safely.

---

# WebView2 runtime and loader

WebGo depends on:

- Microsoft Edge WebView2 Runtime
- `WebView2Loader.dll`

The loader searches common development and deployment locations, including:

- the process working directory
- `webview/loader/WebView2Loader.dll`
- `loader/WebView2Loader.dll`
- the executable directory
- `WebView2Loader.dll` next to the executable

For deployment, the simplest layout is:

```text
MyApplication/
├── MyApplication.exe
└── WebView2Loader.dll
```

The WebView2 Runtime itself must also be installed.

---

# Security notes

Bindings can expose powerful native capabilities.

Treat the web/native bridge as a security boundary.

Recommended rules:

1. Do not navigate privileged WebViews to arbitrary websites.
2. Use `SetNavigationHandler` to restrict allowed origins.
3. Validate every argument received by native bindings.
4. Do not expose arbitrary filesystem or shell execution to untrusted pages.
5. Do not place credentials directly into page JavaScript.
6. Use a dedicated `UserDataFolder` when browser-profile isolation matters.
7. Disable debug features in production unless they are explicitly required.

---

# Typical application model

A WebGo application often looks like:

```text
Go backend
├── database
├── networking
├── filesystem
├── native Windows integration
└── business logic

        │
        │ WebGo Bind / Eval / WebMessage
        ▼

WebView2 frontend
├── HTML
├── CSS
└── JavaScript
```

The frontend handles presentation.

Go handles native and backend functionality.

WebGo connects the two.
