# WebGo Examples

This guide shows practical patterns for building applications with WebGo.

> Examples use the intended package path `github.com/orwf/webgo_repo/webview`. If your checkout currently uses a different local module path, adjust the import accordingly.

---

# 1. Minimal application

```go
package main

import "github.com/orwf/webgo_repo/webview"

func main() {
    w := webview.New(true)
    defer w.Destroy()

    w.SetTitle("WebGo")
    w.SetSize(1000, 700, webview.HintNone)

    w.Navigate("https://example.com")
    w.Run()
}
```

---

# 2. Self-contained HTML application

```go
package main

import "github.com/orwf/webgo_repo/webview"

func main() {
    w := webview.New(true)
    defer w.Destroy()

    w.SetTitle("Embedded HTML")
    w.SetSize(900, 600, webview.HintNone)

    w.NavigateToString(`
<!doctype html>
<html>
<head>
<meta charset="utf-8">
<style>
body {
    margin: 0;
    padding: 40px;
    background: #111418;
    color: #eee;
    font-family: system-ui;
}
button {
    padding: 10px 16px;
}
</style>
</head>
<body>
    <h1>WebGo</h1>
    <p>This page came directly from a Go string.</p>
</body>
</html>
`)

    w.Run()
}
```

---

# 3. Calling Go from JavaScript

Go:

```go
w.Bind("greet", func(name string) string {
    return "Hello, " + name
})
```

HTML/JavaScript:

```html
<button id="hello">Call Go</button>
<pre id="output"></pre>

<script>
document.querySelector("#hello").onclick = async () => {
    const result = await greet("World");
    document.querySelector("#output").textContent = result;
};
</script>
```

---

# 4. Returning structured data

Go:

```go
type SystemInfo struct {
    OS   string `json:"os"`
    Arch string `json:"arch"`
    CPUs int    `json:"cpus"`
}

w.Bind("getSystemInfo", func() SystemInfo {
    return SystemInfo{
        OS: runtime.GOOS,
        Arch: runtime.GOARCH,
        CPUs: runtime.NumCPU(),
    }
})
```

JavaScript:

```js
const info = await getSystemInfo();

console.log(info.os);
console.log(info.arch);
console.log(info.cpus);
```

---

# 5. Passing an object from JavaScript to Go

Go:

```go
type Settings struct {
    Theme string `json:"theme"`
    Zoom  int    `json:"zoom"`
}

w.Bind("saveSettings", func(settings Settings) error {
    fmt.Printf("settings: %+v\n", settings)
    return nil
})
```

JavaScript:

```js
await saveSettings({
    theme: "dark",
    zoom: 110
});
```

---

# 6. Handling errors

Go:

```go
w.Bind("divide", func(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("cannot divide by zero")
    }

    return a / b, nil
})
```

JavaScript:

```js
try {
    const value = await divide(10, 0);
    console.log(value);
} catch (err) {
    console.error("Go error:", err.message);
}
```

---

# 7. Database-style binding

A common WebGo pattern is keeping database access entirely in Go.

```go
type User struct {
    ID       uint64 `json:"id"`
    Username string `json:"username"`
    Admin    bool   `json:"admin"`
}

w.Bind("searchUsers", func(query string) ([]User, error) {
    rows, err := db.Query(
        "SELECT id, username, admin FROM users WHERE username LIKE ? LIMIT 50",
        "%"+query+"%",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User

    for rows.Next() {
        var u User

        if err := rows.Scan(&u.ID, &u.Username, &u.Admin); err != nil {
            return nil, err
        }

        users = append(users, u)
    }

    return users, rows.Err()
})
```

JavaScript:

```js
const users = await searchUsers("alex");

for (const user of users) {
    console.log(user.id, user.username, user.admin);
}
```

This keeps credentials and SQL access out of frontend JavaScript.

---

# 8. Updating the page from Go

```go
w.Eval(`
document.querySelector("#status").textContent = "Connected";
`)
```

A safer pattern for dynamic values is JSON encoding:

```go
value := "Connected"

encoded, _ := json.Marshal(value)

w.Eval(fmt.Sprintf(
    `document.querySelector("#status").textContent = %s;`,
    encoded,
))
```

---

# 9. Reading data from JavaScript

```go
w.EvalWithResult(`
({
    title: document.title,
    url: location.href,
    width: innerWidth,
    height: innerHeight
})
`, func(result string) {
    fmt.Println("Browser state:", result)
})
```

The result is JSON-encoded text returned by WebView2.

---

# 10. Initialization scripts

Initialization scripts are useful when something must exist on every page.

```go
w.Init(`
window.webgoEnvironment = {
    desktop: true,
    platform: "windows"
};
`)
```

Page JavaScript can then read:

```js
console.log(window.webgoEnvironment.platform);
```

---

# 11. Restricting navigation

When bindings expose native functionality, restrict navigation.

```go
w.SetNavigationHandler(func(uri string) bool {
    if uri == "about:blank" {
        return true
    }

    if strings.HasPrefix(uri, "https://app.example/") {
        return true
    }

    return false
})
```

This blocks navigation to other origins before they can become the active page.

---

# 12. Open external links in the system browser

One useful policy is allowing only your application inside WebView2 and forwarding everything else to Windows.

```go
w.SetNavigationHandler(func(uri string) bool {
    if strings.HasPrefix(uri, "https://app.example/") {
        return true
    }

    exec.Command("rundll32", "url.dll,FileProtocolHandler", uri).Start()
    return false
})
```

Validate schemes before launching external URLs in production applications.

---

# 13. Background work without freezing the UI

Bound Go functions are suitable for longer work.

```go
w.Bind("calculateReport", func(input ReportInput) (Report, error) {
    return buildReport(input)
})
```

JavaScript remains Promise-based:

```js
button.disabled = true;

try {
    const report = await calculateReport(input);
    renderReport(report);
} finally {
    button.disabled = false;
}
```

---

# 14. Go background worker updating the UI

```go
go func() {
    for {
        value := readStatus()

        encoded, _ := json.Marshal(value)

        w.Eval(fmt.Sprintf(
            `window.updateStatus(%s);`,
            encoded,
        ))

        time.Sleep(time.Second)
    }
}()
```

Because `Eval` uses WebGo's dispatch path, the JavaScript execution is sent back to the WebView UI thread.

---

# 15. Raw messages from Go to JavaScript

Go:

```go
w.PostMessage("server-connected")
```

JavaScript:

```js
window.chrome.webview.addEventListener("message", event => {
    console.log("native message:", event.data);
});
```

For request/response calls, `Bind` is normally easier. Raw WebMessages are useful for event streams and custom protocols.

---

# 16. Using a dedicated browser profile

```go
w, err := webview.NewWithOptionsE(webview.WebViewOptions{
    Debug: false,
    AutoFocus: true,
    UserDataFolder: "./userdata/account-a",
    Window: webview.WindowOptions{
        Title: "Account A",
        Width: 1100,
        Height: 760,
    },
})
if err != nil {
    log.Fatal(err)
}
```

A separate user-data folder can isolate cookies, storage and cache.

---

# 17. Production constructor

Prefer the error-returning constructor in production:

```go
w, err := webview.NewWithOptionsE(webview.WebViewOptions{
    Debug: false,
    AutoFocus: true,
    Window: webview.WindowOptions{
        Title: "Application",
        Width: 1280,
        Height: 800,
    },
})
if err != nil {
    log.Fatalf("failed to start WebGo: %v", err)
}
defer w.Destroy()
```

---

# 18. Native HWND integration

```go
hwnd := w.HWND()
fmt.Printf("HWND: 0x%X\n", hwnd)
```

The handle can be passed to other Win32 APIs.

Possible uses include:

- custom DWM attributes
- native menus
- taskbar integration
- window placement
- frameless/titlebar work
- hotkeys

---

# 19. Multiple windows

```go
first, err := webview.NewWithOptionsE(webview.WebViewOptions{
    Window: webview.WindowOptions{
        Title: "Window One",
        Width: 800,
        Height: 600,
    },
})
if err != nil {
    log.Fatal(err)
}
defer first.Destroy()

second, err := webview.NewWithOptionsE(webview.WebViewOptions{
    Window: webview.WindowOptions{
        Title: "Window Two",
        Width: 800,
        Height: 600,
    },
})
if err != nil {
    log.Fatal(err)
}
defer second.Destroy()

first.NavigateToString("<h1>First</h1>")
second.NavigateToString("<h1>Second</h1>")

first.Run()
```

WebGo dispatch queues are associated with individual HWNDs, which is required for independent multi-window operation.

Application-level message-loop design may still need to be coordinated when building more complex multi-window applications.

---

# 20. DevTools

```go
w.OpenDevTools()
```

Useful while developing:

- inspect DOM
- debug JavaScript
- inspect requests
- inspect local storage
- profile performance

For production builds, create the WebView with `Debug: false` unless DevTools are intentionally exposed.

---

# 21. A complete mini application

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "strings"

    "github.com/orwf/webgo_repo/webview"
)

type AppInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

func main() {
    w, err := webview.NewWithOptionsE(webview.WebViewOptions{
        Debug: true,
        AutoFocus: true,
        Window: webview.WindowOptions{
            Title: "WebGo Demo",
            Width: 1000,
            Height: 700,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer w.Destroy()

    w.SetNavigationHandler(func(uri string) bool {
        return uri == "about:blank" ||
            strings.HasPrefix(uri, "data:")
    })

    if err := w.Bind("getAppInfo", func() AppInfo {
        return AppInfo{
            Name: "WebGo Demo",
            Version: "1.0",
        }
    }); err != nil {
        log.Fatal(err)
    }

    if err := w.Bind("sum", func(a, b int) int {
        return a + b
    }); err != nil {
        log.Fatal(err)
    }

    w.NavigateToString(`
<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>WebGo Demo</title>
<style>
body {
    background: #101318;
    color: #f3f4f6;
    font: 16px system-ui;
    padding: 32px;
}
button {
    padding: 8px 14px;
}
pre {
    background: #171b22;
    padding: 16px;
    border-radius: 8px;
}
</style>
</head>
<body>
<h1>WebGo Demo</h1>
<button id="load">Call Go</button>
<pre id="output"></pre>

<script>
document.querySelector("#load").onclick = async () => {
    const app = await getAppInfo();
    const total = await sum(20, 22);

    document.querySelector("#output").textContent =
        JSON.stringify({ app, total }, null, 2);
};
</script>
</body>
</html>
`)

    w.Run()

    _, _ = json.Marshal(nil)
    fmt.Print("")
}
```

The unused imports in a real application should of course be removed; they are shown here only to illustrate common packages used in larger examples.
