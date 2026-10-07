//go:build windows && amd64

package main

import (
	"fmt"
	"log"

	webgo "github.com/orwf/webgo_repo"
)

func main() {
	w, err := webgo.NewWithOptionsE(webgo.WebViewOptions{
		Debug:          true,
		AutoFocus:      true,
		UserDataFolder: "./demo-profile",
		Window: webgo.WindowOptions{
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

	w.NavigateToString(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Native Info</title>
<style>
body{background:#11161c;color:#edf2f7;font:16px system-ui;padding:32px}
button,input{padding:9px 12px;margin:5px;font:inherit}
pre{background:#090d11;padding:14px;border-radius:8px}
</style>

<style id="webgo-art-theme">
:root{
  color-scheme:dark;
  --wg-bg:#080808;--wg-panel:#11110f;--wg-panel2:#191916;--wg-line:#34342f;
  --wg-ink:#f1f1ea;--wg-muted:#aaa99e;--wg-acid:#d9ff4f;--wg-coral:#ff665a;
  --wg-blue:#79a7ff;--wg-lav:#b69cff;--wg-cyan:#62e6dc;
  --wg-shadow:0 26px 80px #0009;
}
html,body{background:var(--wg-bg)!important;color:var(--wg-ink)!important}
body{position:relative;min-height:100vh;overflow-x:hidden}
body:before{
  content:"";position:fixed;z-index:-3;inset:0;
  background:
    radial-gradient(circle at 82% -5%,#4b32ff44 0 14%,transparent 31%),
    linear-gradient(151deg,transparent 0 38%,#ff5a4f33 38% 44%,transparent 44% 100%),
    linear-gradient(28deg,transparent 0 72%,#7fffe722 72% 78%,transparent 78% 100%),
    #080808;
}
body:after{
  content:"";position:fixed;z-index:-2;width:430px;height:430px;right:-150px;top:-170px;
  border:56px solid var(--wg-acid);border-radius:50%;opacity:.52;transform:rotate(22deg);
  pointer-events:none;
}
h1,h2,h3{letter-spacing:-.035em}
h1{font-weight:900!important}
button,.btn{
  border-radius:2px!important;
  border:1px solid var(--wg-line)!important;
  background:#1a1a18!important;
  color:var(--wg-ink)!important;
  box-shadow:none!important;
  transition:transform .12s ease,filter .12s ease,background .12s ease!important;
}
button:hover,.btn:hover{transform:translateY(-1px);filter:brightness(1.14)}
button.primary,.btn.primary{background:var(--wg-acid)!important;color:#111!important;border-color:var(--wg-acid)!important}
button.danger,.btn.danger{background:#421d20!important;color:#ffaaa4!important;border-color:#6b292c!important}
input,select,textarea,.input,.select,.textarea{
  border-radius:2px!important;background:#090909!important;color:white!important;border:1px solid #3d3d38!important;
}
.card,.panel,.modal,.column,.task{
  border-radius:2px!important;
  border-color:var(--wg-line)!important;
  box-shadow:var(--wg-shadow);
}
pre{
  border-radius:2px!important;background:#080808!important;border:1px solid #30302c!important;color:#b9cae6!important;
}
table{background:#0e0e0e}
th{color:#8f8f86!important;text-transform:uppercase;letter-spacing:.08em}
a{color:var(--wg-blue)}
</style>
</head>
<body>
<div aria-hidden="true" style="position:fixed;z-index:-1;left:-110px;top:38%;width:420px;height:78px;background:#ff665a99;transform:rotate(-31deg);pointer-events:none"></div>
<div aria-hidden="true" style="position:fixed;z-index:-1;right:24%;top:13%;width:170px;height:470px;background:linear-gradient(#62e6dc00,#62e6dc40,#62e6dc00);transform:rotate(14deg);pointer-events:none"></div>
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
</html>`)

	w.Run()
}
