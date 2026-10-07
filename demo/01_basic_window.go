//go:build windows && amd64

package main

import (
	"log"

	"webgo_repo-main/webview"
)

func main() {
	w, err := webview.NewWithOptionsE(webview.WebViewOptions{
		Debug:     true,
		AutoFocus: true,
		Window: webview.WindowOptions{
			Title:  "WebGo Demo 01 - Basic Window",
			Width:  1000,
			Height: 700,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer w.Destroy()

	w.NavigateToString(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>WebGo Basic Window</title>
<style>
body{margin:0;background:#0f1318;color:#edf2f7;font:16px system-ui;padding:40px}
.card{max-width:760px;margin:auto;padding:28px;border:1px solid #2c3642;border-radius:14px;background:#171d24}
code{background:#0b0e12;padding:3px 7px;border-radius:6px}
</style>
</head>
<body>
<div class="card">
<h1>WebGo is running</h1>
<p>This page was loaded directly from a Go string with <code>NavigateToString</code>.</p>
<p>The native window, WebView2 renderer and Go backend are all active.</p>
</div>
</body>
</html>`)

	w.Run()
}
