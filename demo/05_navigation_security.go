//go:build windows && amd64

package main

import (
	"log"
	"strings"

	webgo "github.com/orwf/webgo_repo"
)

func main() {
	w := webgo.New(true)
	defer w.Destroy()

	w.SetTitle("WebGo Demo 05 - Navigation Security")
	w.SetSize(1050, 720, webgo.HintNone)

	w.SetNavigationHandler(func(uri string) bool {
		if uri == "about:blank" {
			return true
		}
		if strings.HasPrefix(uri, "data:") {
			return true
		}
		if strings.HasPrefix(uri, "https://example.com") {
			return true
		}
		return false
	})

	if err := w.Bind("nativeSecret", func() string {
		return "This represents privileged native functionality."
	}); err != nil {
		log.Fatal(err)
	}

	w.NavigateToString(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Navigation Policy</title>
<style>
body{background:#10151b;color:#eee;font:16px system-ui;padding:32px}
button{padding:9px 13px;margin:5px}
code{background:#070a0d;padding:3px 6px}
</style>
</head>
<body>
<h1>Navigation policy demo</h1>
<p>This WebView allows only <code>about:blank</code>, data pages and <code>https://example.com</code>.</p>
<button onclick="location.href='https://example.com'">Allowed navigation</button>
<button onclick="location.href='https://github.com'">Blocked navigation</button>
<button id="secret">Call native binding</button>
<p id="out"></p>
<script>
document.querySelector("#secret").onclick = async () => {
  document.querySelector("#out").textContent = await nativeSecret();
};
</script>
</body>
</html>`)

	w.Run()
}
