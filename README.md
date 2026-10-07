# WebGo

A lightweight native Windows WebView2 framework for Go.

WebGo provides direct access to Microsoft Edge WebView2 from Go using Win32 and COM, with a built-in asynchronous JavaScript ↔ Go bridge for building desktop applications with a web frontend and native Go backend.

The project is designed to provide a smaller, more direct alternative to Electron-style desktop application stacks while retaining the flexibility of HTML, CSS, and JavaScript for the user interface.

> **Platform:** Windows  
> **Backend:** Go  
> **Frontend:** HTML / CSS / JavaScript  
> **Renderer:** Microsoft Edge WebView2  
> **Native API:** Win32 + COM

---

## Features

- Native Windows application windows
- Microsoft Edge WebView2 rendering
- Go ↔ JavaScript function bindings
- Promise-based JavaScript calls into Go
- Asynchronous Go binding execution
- JavaScript evaluation from Go
- JavaScript evaluation with results
- WebView message handling
- Navigation events
- Document initialization scripts
- WebView2 DevTools support
- Dynamic window resizing
- Custom window titles
- Native Win32 message loop
- WebView2 runtime detection
- WebView2Loader support
- No Electron runtime
- Low-level WebView2 access
- Designed for embedding native Go services behind modern web interfaces


---

# Documentation

WebGo now includes dedicated documentation for both application developers and contributors:

| Guide | What it covers |
| --- | --- |
| [API Reference](docs/API.md) | Every public WebView method, options, navigation, bindings, JavaScript execution, sizing, lifecycle, HWND access and security notes |
| [Examples](docs/EXAMPLES.md) | Complete practical examples for bindings, structured data, errors, databases, navigation rules, background work, multiple windows and native integration |
| [Architecture](docs/ARCHITECTURE.md) | WebView2 initialization, COM interfaces, callbacks, GC anchors, dispatch queues, Win32 message handling, bridge internals and destruction order |

## Public API at a glance

| Capability | API |
| --- | --- |
| Create a WebView | `New`, `NewWithOptions`, `NewWithOptionsE` |
| Start/stop lifecycle | `Run`, `Terminate`, `Destroy` |
| UI-thread execution | `Dispatch` |
| Window title | `SetTitle` |
| Window size and constraints | `SetSize` with `HintNone`, `HintFixed`, `HintMin`, `HintMax` |
| Navigate to a URL | `Navigate` |
| Load HTML directly | `NavigateToString` |
| Filter/cancel navigation | `SetNavigationHandler` |
| Install document-start JavaScript | `Init` |
| Execute JavaScript | `Eval`, `EvalDirect` |
| Execute JavaScript and read the result | `EvalWithResult` |
| Expose Go functions to JavaScript | `Bind`, `Unbind` |
| Send raw native-to-page messages | `PostMessage` |
| Open Edge DevTools | `OpenDevTools` |
| Read WebView2 runtime version | `Version` |
| Access the native HWND | `HWND` |

## What you can build with WebGo

The current feature set is suitable for Windows applications where a web frontend needs access to native Go logic, including:

- desktop dashboards and administration tools
- SQL/database frontends
- TCP, Telnet or WebSocket clients
- launchers and configuration utilities
- local server control panels
- file-management tools
- monitoring and diagnostics applications
- native applications with HTML/CSS interfaces
- tools that need direct Win32 access alongside a modern UI
- applications that isolate browser state with separate WebView2 user-data folders

The JavaScript bridge is Promise based, so native Go operations can be consumed from frontend code using ordinary `async` / `await`.

For concrete code, start with the [Examples guide](docs/EXAMPLES.md).

# Architecture

WebGo sits between a native Go application and WebView2.

```text
┌───────────────────────────────────────┐
│             Go Application            │
│                                       │
│   Networking / Database / Files       │
│   Application Logic / Native APIs     │
└──────────────────┬────────────────────┘
                   │
                   │ Go bindings
                   │
┌──────────────────▼────────────────────┐
│                 WebGo                 │
│                                       │
│   Window Management                   │
│   Go ↔ JavaScript Bridge              │
│   WebView2 / COM                      │
│   Win32                               │
└──────────────────┬────────────────────┘
                   │
                   │ WebView2
                   │
┌──────────────────▼────────────────────┐
│             Web Frontend              │
│                                       │
│        HTML / CSS / JavaScript        │
└───────────────────────────────────────┘
```

JavaScript can call registered Go functions and receive their result through a Promise.

This makes it possible to keep application logic in Go while building the interface using standard web technologies.

---

# Project Structure

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
    └── w32/
```

### `webview.go`

Provides the main public WebGo API.

It manages:

- WebView creation
- navigation
- JavaScript evaluation
- Go function bindings
- JavaScript message handling
- dispatching
- application lifecycle

### `edge/`

Contains the WebView2 implementation.

This includes:

- WebView2 COM interfaces
- environment creation
- controller creation
- browser initialization
- navigation callbacks
- JavaScript callbacks
- WebMessage handling

### `loader/`

Handles loading `WebView2Loader.dll` and creating the WebView2 environment.

### `wAPI/`

Contains the lower-level Windows implementation used by WebGo.

This includes:

- Win32 window creation
- Windows messages
- COM helpers
- dispatching
- native handles
- Windows API definitions

---

# Requirements

WebGo currently targets **Windows**.

You will need:

- Windows 10 or Windows 11
- Go
- Microsoft Edge WebView2 Runtime
- `WebView2Loader.dll`

Most modern Windows systems already have the WebView2 Runtime installed.

Microsoft applications and many Windows desktop applications also install it automatically.

---

# WebView2Loader.dll

WebGo uses Microsoft's native:

```text
WebView2Loader.dll
```

During development the repository currently contains the loader under:

```text
webview/loader/WebView2Loader.dll
```

The loader attempts to locate the DLL from supported development/application locations.

For release builds, the recommended application layout is:

```text
MyApplication/
├── MyApplication.exe
└── WebView2Loader.dll
```

Keeping the DLL beside the executable avoids depending on the process working directory.

---

# Basic Usage

Create a WebView:

```go
package main

import (
	"your-module/webview"
)

func main() {
	w := webview.New(true)

	defer w.Destroy()

	w.SetTitle("WebGo Application")
	w.SetSize(1200, 800)

	w.Navigate("https://example.com")

	w.Run()
}
```

The exact import path should be replaced with the module path used by your project.

---

# Loading HTML

WebGo can also be used as the native backend for an HTML/CSS/JavaScript application.

For example:

```go
html := `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">

	<title>WebGo</title>

	<style>
		body {
			background: #101114;
			color: #ffffff;
			font-family: system-ui;
			padding: 40px;
		}
	</style>
</head>

<body>

	<h1>WebGo</h1>

	<p>Running inside WebView2.</p>

</body>
</html>
`

w.NavigateToString(html)
```

---

# Calling Go from JavaScript

One of the main features of WebGo is its JavaScript ↔ Go bridge.

A Go function can be exposed to the frontend using `Bind`.

For example:

```go
err := w.Bind("greet", func(name string) string {
	return "Hello " + name
})

if err != nil {
	panic(err)
}
```

The function becomes available to JavaScript.

```js
const result = await window.greet("World");

console.log(result);
```

Result:

```text
Hello World
```

Bindings use JavaScript Promises, allowing Go operations to integrate naturally with asynchronous frontend code.

---

# Binding Functions That Return Errors

Go functions can return an error:

```go
w.Bind("divide", func(a float64, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}

	return a / b, nil
})
```

JavaScript can handle the failure through the Promise:

```js
try {
	const result = await divide(10, 0);

	console.log(result);
} catch (err) {
	console.error(err);
}
```

---

# Native Backend Operations

Bindings can be used to expose native Go functionality to the frontend.

For example:

```go
w.Bind("getSystemInfo", func() map[string]interface{} {
	return map[string]interface{}{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"cpu":  runtime.NumCPU(),
	}
})
```

JavaScript:

```js
const system = await getSystemInfo();

console.log(system.os);
console.log(system.arch);
console.log(system.cpu);
```

This architecture can be used for operations such as:

- database queries
- TCP connections
- filesystem operations
- configuration management
- local services
- system information
- application state
- native Windows functionality

---

# Long-Running Go Bindings

WebGo executes bound application functions asynchronously so long-running backend operations do not need to block WebView2's UI callback.

For example:

```go
w.Bind("searchDatabase", func(query string) ([]Result, error) {
	return database.Search(query)
})
```

The frontend can simply use:

```js
const results = await searchDatabase(searchInput);
```

This is particularly useful for database, network, and filesystem operations.

---

# Executing JavaScript From Go

JavaScript can be executed from the Go backend:

```go
w.Eval(`
	document.body.dataset.connected = "true";
`)
```

This can be used to update the frontend in response to native application events.

---

# Getting JavaScript Results

WebGo also supports asynchronous JavaScript evaluation with a result:

```go
w.EvalWithResult(
	`document.title`,
	func(result string) {
		fmt.Println("Document title:", result)
	},
)
```

The completion callback is retained until WebView2 completes the asynchronous operation.

---

# Initialization Scripts

JavaScript that must exist when a document is created can be registered as an initialization script.

```go
w.Init(`
	window.webgo = {
		version: "development"
	};
`)
```

Initialization scripts are useful for:

- frontend APIs
- application configuration
- bridge initialization
- environment information
- functions required before page scripts execute

---

# Developer Tools

WebView2 developer tools can be opened from Go:

```go
w.OpenDevTools()
```

This gives access to the normal Chromium/Edge development environment, including:

- Console
- Elements
- Network
- Sources
- Performance
- Storage

---

# Dispatching to the UI Thread

WebView2 and Win32 operations often need to execute on the window/UI thread.

WebGo provides dispatching for this purpose:

```go
w.Dispatch(func() {
	w.SetTitle("Updated from Go")
})
```

Application code performing work in goroutines should dispatch UI-specific work back to the WebView thread when required.

---

# Example Application Architecture

A larger WebGo application can use a structure such as:

```text
myapp/
├── main.go
│
├── backend/
│   ├── database.go
│   ├── network.go
│   └── settings.go
│
├── frontend/
│   ├── index.html
│   ├── app.js
│   └── style.css
│
└── WebView2Loader.dll
```

Go handles the application backend:

```text
Database
Network
Filesystem
Native APIs
Application state
```

while WebView2 handles:

```text
HTML
CSS
JavaScript
UI rendering
```

The WebGo bridge connects the two.

---

# Security

A WebView with native Go bindings should be treated as a security boundary.

A bound Go function may potentially expose access to:

- files
- databases
- network connections
- operating-system functions
- application credentials
- other native resources

Do not expose privileged bindings to arbitrary or untrusted remote pages.

Applications should restrict navigation when native bindings provide sensitive functionality.

Inputs received through JavaScript bindings should also be validated by Go.

Do not assume input is trusted simply because it originated from the application's frontend.

---

# Current Status

WebGo is under active development.

The project currently provides the core infrastructure required to create Windows WebView2 applications directly from Go.

Areas still being developed or hardened include:

- COM callback/reference lifetime handling
- multi-window dispatching
- navigation policies
- richer WebView2 event support
- resource interception
- virtual host mapping
- downloads
- permission handling
- custom browser profiles
- crash/process recovery
- additional WebView2 settings
- testing
- examples
- packaging

The public API may change while these systems are developed.

---

# Roadmap

Planned areas of development include:

### Core

- stronger COM lifetime management
- per-window dispatch queues
- improved error propagation
- deterministic shutdown
- thread-safe lifecycle management

### JavaScript Bridge

- versioned RPC protocol
- structured errors
- request cancellation
- Go → JavaScript events
- typed frontend APIs
- request timeouts

### WebView2

- navigation filtering
- resource interception
- virtual host mapping
- downloads
- permission requests
- new-window handling
- process failure handling
- custom user agents
- profiles
- cookies
- browser settings

### Native Windows

- file dialogs
- frameless windows
- title-bar customization
- clipboard integration
- drag and drop
- system tray support
- native menus

### Development

- examples
- unit tests
- integration tests
- Windows CI
- automated builds
- versioned releases

---

# Design Goals

WebGo aims to remain:

**Small**

Avoid shipping an entire browser runtime when Windows already provides WebView2.

**Native**

Use Go, Win32, COM, and WebView2 directly.

**Flexible**

Allow standard HTML, CSS, and JavaScript to be used for desktop interfaces.

**Go-first**

Keep application logic, networking, databases, concurrency, and native functionality in Go.

**Low-level when necessary**

WebGo does not attempt to hide every native concept. Access to WebView2 and Windows functionality should remain possible when applications need it.

---

# Why WebGo?

Traditional web-based desktop frameworks often bundle a substantial runtime with every application.

WebGo takes a different approach:

```text
Electron-style application

Application
    +
Node.js
    +
Chromium
    +
Framework runtime


WebGo application

Application
    +
Go
    +
Windows WebView2
```

WebGo uses the WebView2 runtime already available on modern Windows systems while Go provides the native backend.

The result is intended to be a compact architecture for applications that need both a modern web interface and native Go functionality.

---

# Contributing

WebGo is still evolving and contributions are welcome.

Useful areas for contributions include:

- WebView2 interface coverage
- COM correctness
- Win32 integration
- concurrency and lifecycle handling
- tests
- examples
- documentation
- debugging tools
- packaging

When contributing low-level COM or Win32 changes, test application startup, navigation, JavaScript bindings, window resizing, shutdown, and repeated creation/destruction carefully.

---

# License

Add the project's license here.

If the repository is intended to be open source, add a `LICENSE` file to the repository and update this section with the selected license.

---

# Acknowledgements

WebGo uses Microsoft's WebView2 platform to host the Microsoft Edge rendering engine inside native Windows applications.

WebView2 and Microsoft Edge are products of Microsoft.

---

## Project Status

**Experimental / Active Development**

WebGo is usable as a foundation for Windows WebView2 applications, but its lower-level APIs and internal architecture may continue to change as COM lifecycle management, multi-window support, event handling, and the JavaScript bridge are expanded.
