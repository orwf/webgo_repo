//go:build windows && amd64

package main

import (
	"log"

	webgo "github.com/orwf/webgo_repo"
)

func main() {
	w := webgo.New(true)
	defer w.Destroy()

	w.SetTitle("WebGo Demo 06 - Window Sizing")
	w.SetSize(1000, 700, webgo.HintNone)

	if err := w.Bind("normalSize", func() string {
		w.SetSize(1000, 700, webgo.HintNone)
		return "Normal resizing enabled."
	}); err != nil {
		log.Fatal(err)
	}

	if err := w.Bind("fixedSize", func() string {
		w.SetSize(900, 600, webgo.HintFixed)
		return "Window locked to 900x600."
	}); err != nil {
		log.Fatal(err)
	}

	if err := w.Bind("minimumSize", func() string {
		w.SetSize(700, 500, webgo.HintMin)
		return "Minimum size set to 700x500."
	}); err != nil {
		log.Fatal(err)
	}

	if err := w.Bind("maximumSize", func() string {
		w.SetSize(1200, 850, webgo.HintMax)
		return "Maximum size set to 1200x850."
	}); err != nil {
		log.Fatal(err)
	}

	w.NavigateToString(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Sizing</title>
<style>
body{background:#11161c;color:#eee;font:16px system-ui;padding:32px}
button{padding:10px 13px;margin:6px}
#out{margin-top:20px;font-weight:600}
</style>
</head>
<body>
<h1>Window sizing hints</h1>
<button onclick="apply(normalSize)">HintNone</button>
<button onclick="apply(fixedSize)">HintFixed</button>
<button onclick="apply(minimumSize)">HintMin</button>
<button onclick="apply(maximumSize)">HintMax</button>
<div id="out">Choose a mode, then try resizing the window.</div>
<script>
async function apply(fn) {
  document.querySelector("#out").textContent = await fn();
}
</script>
</body>
</html>`)

	w.Run()
}
