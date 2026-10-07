//go:build windows && amd64

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"webgo_repo-main/webview"
)

func main() {
	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("WebGo Demo 04 - Eval")
	w.SetSize(1000, 700, webview.HintNone)

	if err := w.Bind("requestGoEval", func() string {
		go func() {
			time.Sleep(250 * time.Millisecond)
			w.Eval(`
document.querySelector("#from-go").textContent =
  "This text was changed by Go using Eval().";
`)
		}()
		return "Go scheduled an Eval() call."
	}); err != nil {
		log.Fatal(err)
	}

	if err := w.Bind("inspectPage", func() string {
		w.EvalWithResult(`({
  title: document.title,
  url: location.href,
  width: innerWidth,
  height: innerHeight
})`, func(result string) {
			fmt.Println("EvalWithResult:", result)

			message := "Raw WebView2 JSON result: " + result
			encoded, _ := json.Marshal(message)

			w.Eval(fmt.Sprintf(
				`document.querySelector("#result").textContent = %s;`,
				encoded,
			))
		})
		return "Inspection requested."
	}); err != nil {
		log.Fatal(err)
	}

	w.NavigateToString(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Eval Demo</title>
<style>
body{background:#11161c;color:#eef;font:16px system-ui;padding:35px}
button{padding:10px 14px;margin:5px}
pre{background:#090d11;padding:15px;border-radius:8px}
</style>
</head>
<body>
<h1>Go → JavaScript execution</h1>
<p id="from-go">Waiting for Eval().</p>
<button id="eval">Ask Go to call Eval()</button>
<button id="inspect">Ask Go to call EvalWithResult()</button>
<pre id="status">Ready.</pre>
<pre id="result">No result yet.</pre>
<script>
document.querySelector("#eval").onclick = async () => {
  document.querySelector("#status").textContent = await requestGoEval();
};
document.querySelector("#inspect").onclick = async () => {
  document.querySelector("#status").textContent = await inspectPage();
};
</script>
</body>
</html>`)

	w.Run()
}
