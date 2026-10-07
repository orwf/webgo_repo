//go:build windows && amd64

package main

import (
	"log"
	"runtime"

	"github.com/orwf/webgo_repo/webview"
)

type SystemInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	CPUs int    `json:"cpus"`
}

func main() {
	w := webgo.New(true)
	defer w.Destroy()

	w.SetTitle("WebGo Demo 02 - Go Bindings")
	w.SetSize(1000, 700, webgo.HintNone)

	if err := w.Bind("greet", func(name string) string {
		return "Hello, " + name + " — this came from Go."
	}); err != nil {
		log.Fatal(err)
	}

	if err := w.Bind("add", func(a, b int) int {
		return a + b
	}); err != nil {
		log.Fatal(err)
	}

	if err := w.Bind("getSystemInfo", func() SystemInfo {
		return SystemInfo{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
			CPUs: runtime.NumCPU(),
		}
	}); err != nil {
		log.Fatal(err)
	}

	w.NavigateToString(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Bindings</title>
<style>
body{background:#11151b;color:#f5f7fa;font:16px system-ui;padding:32px}
button,input{font:inherit;padding:9px 12px;margin:5px}
pre{background:#090c10;padding:16px;border-radius:10px;white-space:pre-wrap}
</style>
</head>
<body>
<h1>JavaScript → Go bindings</h1>
<input id="name" value="World">
<button id="greet">Call greet()</button>
<button id="math">Call add(20, 22)</button>
<button id="system">Get system info</button>
<pre id="out">Ready.</pre>
<script>
const out = document.querySelector("#out");
document.querySelector("#greet").onclick = async () => {
  out.textContent = await greet(document.querySelector("#name").value);
};
document.querySelector("#math").onclick = async () => {
  out.textContent = "20 + 22 = " + await add(20, 22);
};
document.querySelector("#system").onclick = async () => {
  out.textContent = JSON.stringify(await getSystemInfo(), null, 2);
};
</script>
</body>
</html>`)

	w.Run()
}
