//go:build windows && amd64

package main

import (
	"encoding/json"
	"log"
	"time"

	"webgo_repo-main/webview"
)

func main() {
	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("WebGo Demo 07 - Init and Messages")
	w.SetSize(1000, 700, webview.HintNone)

	w.Init(\`
window.demoBoot = {
  installedByGo: true,
  loadedAt: new Date().toISOString()
};

window.chrome.webview.addEventListener("message", event => {
  const el = document.querySelector("#native-events");
  if (!el) return;
  el.textContent += "\\n" + event.data;
});
\`)

	if err := w.Bind("startNativeEvents", func() string {
		go func() {
			for i := 1; i <= 5; i++ {
				payload, _ := json.Marshal(map[string]interface{}{
					"type":  "tick",
					"count": i,
					"time":  time.Now().Format(time.RFC3339),
				})

				w.PostMessage(string(payload))
				time.Sleep(time.Second)
			}
		}()

		return "Started five native WebMessage events."
	}); err != nil {
		log.Fatal(err)
	}

	w.NavigateToString(\`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Init + Messages</title>
<style>
body{background:#10151a;color:#eee;font:16px system-ui;padding:30px}
button{padding:9px 14px}
pre{background:#070b0e;padding:14px;border-radius:8px}
</style>
</head>
<body>
<h1>Init scripts + raw WebMessages</h1>
<pre id="boot"></pre>
<button id="start">Start native events</button>
<pre id="status">Ready.</pre>
<pre id="native-events">Messages:</pre>
<script>
document.querySelector("#boot").textContent =
  "Init object: " + JSON.stringify(window.demoBoot, null, 2);

document.querySelector("#start").onclick = async () => {
  document.querySelector("#status").textContent =
    await startNativeEvents();
};
</script>
</body>
</html>\`)

	w.Run()
}
