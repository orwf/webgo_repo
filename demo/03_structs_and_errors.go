//go:build windows && amd64

package main

import (
	"errors"
	"log"
	"strings"

	"github.com/orwf/webgo_repo/webview"
)

type UserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type UserResult struct {
	Valid   bool     `json:"valid"`
	Summary string   `json:"summary"`
	Tags    []string `json:"tags"`
}

func main() {
	w := webgo.New(true)
	defer w.Destroy()

	w.SetTitle("WebGo Demo 03 - Structs and Errors")
	w.SetSize(1000, 720, webgo.HintNone)

	if err := w.Bind("validateUser", func(in UserInput) (UserResult, error) {
		if strings.TrimSpace(in.Name) == "" {
			return UserResult{}, errors.New("name is required")
		}
		if !strings.Contains(in.Email, "@") {
			return UserResult{}, errors.New("email must contain @")
		}
		if in.Age < 18 {
			return UserResult{}, errors.New("user must be at least 18")
		}

		return UserResult{
			Valid:   true,
			Summary: in.Name + " <" + in.Email + ">",
			Tags:    []string{"validated", "adult"},
		}, nil
	}); err != nil {
		log.Fatal(err)
	}

	w.NavigateToString(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Structs and Errors</title>
<style>
body{background:#101419;color:#eee;font:16px system-ui;padding:32px}
input,button{display:block;width:360px;max-width:100%;margin:8px 0;padding:9px;font:inherit}
button{width:auto}
.ok{color:#8ee6a1}
.bad{color:#ff8e8e}
pre{background:#090c0f;padding:14px;border-radius:8px}
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
<h1>Structured arguments + Go errors</h1>
<input id="name" value="Alice" placeholder="Name">
<input id="email" value="alice@example.com" placeholder="Email">
<input id="age" value="25" type="number" placeholder="Age">
<button id="go">Validate in Go</button>
<pre id="out">Ready.</pre>
<script>
const out = document.querySelector("#out");
document.querySelector("#go").onclick = async () => {
  try {
    const result = await validateUser({
      name: document.querySelector("#name").value,
      email: document.querySelector("#email").value,
      age: Number(document.querySelector("#age").value)
    });
    out.className = "ok";
    out.textContent = JSON.stringify(result, null, 2);
  } catch (err) {
    out.className = "bad";
    out.textContent = "Go rejected the request: " + err.message;
  }
};
</script>
</body>
</html>`)

	w.Run()
}
