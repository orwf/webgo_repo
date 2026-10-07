//go:build windows && amd64

package main

import (
	"fmt"
	"log"

	"webgo_repo-main/webview"
)

func main() {
	w, err := webview.NewWithOptionsE(webview.WebViewOptions{
		Debug:          true,
		AutoFocus:      true,
		UserDataFolder: "./demo-profile",
		Window: webview.WindowOptions{
			Title:  "WebGo Demo 08 - Native Info",
			Width:  1050,
			Height: 720,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer w.Destroy()

	if err := w.Bind("getNativeInfo", func() map[string]string {
		return map[string]string{
			"webview2Version": w.Version(),
			"hwnd":            fmt.Sprintf("0x%X", w.HWND()),
			"profile":         "./demo-profile",
		}
	}); err != nil {
		log.Fatal(err)
	}

	if err := w.Bind("openDevTools", func() string {
		w.OpenDevTools()
		return "DevTools requested."
	}); err != nil {
		log.Fatal(err)
	}

	if err := w.Bind("renameWindow", func(title string) string {
		w.SetTitle(title)
		return "Window title updated."
	}); err != nil {
		log.Fatal(err)
	}

	w.NavigateToString(\`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Native Info</title>
<style>
body{background:#11161c;color:#edf2f7;font:16px system-ui;padding:32px}
button,input{padding:9px 12px;margin:5px;font:inherit}
pre{background:#090d11;padding:14px;border-radius:8px}
</style>
</head>
<body>
<h1>Native WebGo information</h1>
<button id="info">Get HWND + WebView2 version</button>
<button id="dev">Open DevTools</button>
<br>
<input id="title" value="Renamed by WebGo">
<button id="rename">Rename native window</button>
<pre id="out">Ready.</pre>
<script>
const out = document.querySelector("#out");

document.querySelector("#info").onclick = async () => {
  out.textContent = JSON.stringify(await getNativeInfo(), null, 2);
};

document.querySelector("#dev").onclick = async () => {
  out.textContent = await openDevTools();
};

document.querySelector("#rename").onclick = async () => {
  out.textContent = await renameWindow(
    document.querySelector("#title").value
  );
};
</script>
</body>
</html>\`)

	w.Run()
}
