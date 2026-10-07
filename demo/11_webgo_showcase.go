//go:build windows && amd64

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	webgo "github.com/orwf/webgo_repo"
)

type ShowcaseInput struct {
	Name  string  `json:"name"`
	Count int     `json:"count"`
	Rate  float64 `json:"rate"`
}

type ShowcaseResult struct {
	Message string    `json:"message"`
	Values  []float64 `json:"values"`
	Total   float64   `json:"total"`
}

func main() {
	w, err := webgo.NewWithOptionsE(webgo.WebViewOptions{
		Debug: true,
		AutoFocus: true,
		Window: webgo.WindowOptions{
			Title: "WebGo Comprehensive Showcase",
			Width: 1400,
			Height: 880,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer w.Destroy()

	w.SetNavigationHandler(func(uri string) bool {
		return uri == "about:blank" ||
			strings.HasPrefix(uri, "data:") ||
			strings.HasPrefix(uri, "https://example.com")
	})

	w.Init(`
window.showcaseBoot = {
  injectedByGo: true,
  bootedAt: new Date().toISOString()
};
`)

	mustBind := func(name string, fn interface{}) {
		if err := w.Bind(name, fn); err != nil {
			log.Fatal(err)
		}
	}

	mustBind("showHello", func(name string) string {
		name = strings.TrimSpace(name)
		if name == "" {
			name = "WebGo"
		}
		return "Hello " + name + " from native Go"
	})

	mustBind("showCalculate", func(in ShowcaseInput) (ShowcaseResult, error) {
		if in.Count < 1 || in.Count > 20 {
			return ShowcaseResult{}, errors.New("count must be between 1 and 20")
		}
		if math.IsNaN(in.Rate) || math.IsInf(in.Rate, 0) {
			return ShowcaseResult{}, errors.New("rate must be a finite number")
		}

		values := make([]float64, 0, in.Count)
		var total float64
		for i := 1; i <= in.Count; i++ {
			v := float64(i) * in.Rate
			values = append(values, v)
			total += v
		}

		return ShowcaseResult{
			Message: fmt.Sprintf("Generated %d values for %s", in.Count, in.Name),
			Values:  values,
			Total:   total,
		}, nil
	})

	mustBind("showForceError", func(message string) error {
		if strings.TrimSpace(message) == "" {
			message = "intentional Go error"
		}
		return errors.New(message)
	})

	mustBind("showPanicRecovery", func() string {
		panic("intentional demo panic: WebGo should return this as a Promise error")
	})

	mustBind("showRuntimeInfo", func() map[string]interface{} {
		return map[string]interface{}{
			"webview2Version": w.Version(),
			"hwnd":            fmt.Sprintf("0x%X", w.HWND()),
			"time":            time.Now().Format(time.RFC3339),
			"bootMode":        "WebGo root import package",
		}
	})

	mustBind("showRename", func(title string) string {
		title = strings.TrimSpace(title)
		if title == "" {
			title = "WebGo Comprehensive Showcase"
		}
		w.SetTitle(title)
		return "Native window title changed"
	})

	mustBind("showResize", func(mode string) string {
		switch mode {
		case "normal":
			w.SetSize(1200, 760, webgo.HintNone)
			return "Normal 1200x760 window"
		case "fixed":
			w.SetSize(1000, 700, webgo.HintFixed)
			return "Fixed 1000x700 window"
		case "minimum":
			w.SetSize(760, 560, webgo.HintMin)
			return "Minimum size set to 760x560"
		case "maximum":
			w.SetSize(1500, 950, webgo.HintMax)
			return "Maximum size set to 1500x950"
		default:
			return "Unknown resize mode"
		}
	})

	mustBind("showOpenDevTools", func() string {
		w.OpenDevTools()
		return "DevTools requested"
	})

	mustBind("showEvalFromGo", func(text string) string {
		data, _ := json.Marshal("Go changed this element: " + text)
		w.Eval(fmt.Sprintf(
			`document.querySelector("#evalTarget").textContent = %s; document.querySelector("#evalTarget").classList.add("flash"); setTimeout(function(){document.querySelector("#evalTarget").classList.remove("flash")},700);`,
			data,
		))
		return "Eval dispatched to the WebView UI thread"
	})

	mustBind("showInspectPage", func() string {
		w.EvalWithResult(`({
  title: document.title,
  href: location.href,
  width: innerWidth,
  height: innerHeight,
  buttons: document.querySelectorAll("button").length,
  boot: window.showcaseBoot
})`, func(result string) {
			payload, _ := json.Marshal("EvalWithResult => " + result)
			w.Eval(fmt.Sprintf(`document.querySelector("#inspectOut").textContent = %s;`, payload))
		})
		return "Inspection started asynchronously"
	})

	mustBind("showStartEvents", func() string {
		go func() {
			for i := 1; i <= 10; i++ {
				payload, _ := json.Marshal(map[string]interface{}{
					"type":  "native-event",
					"index": i,
					"time":  time.Now().Format("15:04:05"),
				})
				w.PostMessage(string(payload))
				time.Sleep(600 * time.Millisecond)
			}
		}()
		return "Started 10 native WebMessage events"
	})

	mustBind("showBackgroundJob", func(seconds int) (map[string]interface{}, error) {
		if seconds < 1 || seconds > 8 {
			return nil, errors.New("duration must be between 1 and 8 seconds")
		}
		start := time.Now()
		time.Sleep(time.Duration(seconds) * time.Second)
		return map[string]interface{}{
			"durationRequested": seconds,
			"elapsedMs":         time.Since(start).Milliseconds(),
			"completedAt":       time.Now().Format(time.RFC3339),
		}, nil
	})

	mustBind("showNavigateAllowed", func() string {
		w.Navigate("https://example.com")
		return "Navigating to allowed origin"
	})

	mustBind("showNavigateBlocked", func() string {
		w.Navigate("https://github.com")
		return "Navigation was requested, but the navigation policy should cancel it"
	})

	w.NavigateToString(showcaseHTML)
	w.Run()
}

const showcaseHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>WebGo Comprehensive Showcase</title>
<style>
:root{color-scheme:dark;--bg:#090d12;--panel:#111820;--panel2:#17212b;--line:#293746;--ink:#eef4f8;--muted:#8ea0b2;--blue:#5b8cff;--green:#4fd58e;--amber:#ffc15b;--red:#f46f6f;--purple:#a98cff}
*{box-sizing:border-box}body{margin:0;background:radial-gradient(circle at 50% -20%,#1b2d45,#090d12 40%);color:var(--ink);font:14px system-ui,Segoe UI,sans-serif}
header{position:sticky;top:0;z-index:5;display:flex;justify-content:space-between;align-items:center;padding:15px 22px;background:#0b1118ee;border-bottom:1px solid var(--line);backdrop-filter:blur(14px)}
.brand{font-size:20px;font-weight:800}.brand span{color:var(--blue)}.native{color:var(--muted);font-size:12px}.layout{display:grid;grid-template-columns:230px 1fr;min-height:calc(100vh - 58px)}
nav{padding:18px;border-right:1px solid var(--line);background:#0d131a}.tab{display:block;width:100%;border:0;background:transparent;color:var(--muted);text-align:left;padding:10px 12px;border-radius:8px;margin-bottom:4px;cursor:pointer}.tab.active,.tab:hover{background:#18232f;color:white}
main{padding:22px;min-width:0}.view{display:none}.view.active{display:block}.hero{padding:22px;background:linear-gradient(135deg,#162435,#111922);border:1px solid var(--line);border-radius:13px;margin-bottom:16px}.hero h1{margin:0 0 7px;font-size:28px}.muted{color:var(--muted)}
.grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px}.panel{background:#10171f;border:1px solid var(--line);border-radius:12px;padding:16px}.panel h2{font-size:16px;margin:0 0 10px}.panel p{color:var(--muted);line-height:1.45}
button,input,select{font:inherit}.btn{border:1px solid var(--line);background:#18232e;color:white;padding:9px 12px;border-radius:8px;cursor:pointer;margin:3px}.btn:hover{filter:brightness(1.15)}.btn.primary{background:var(--blue);border-color:transparent}.btn.red{background:#482328;color:#ffaaaa}.btn.green{background:#183a2b;color:#91efb7}
.input,.select{background:#090e13;border:1px solid var(--line);color:white;padding:8px 10px;border-radius:7px;margin:3px;outline:none}.input{min-width:220px}
pre{background:#080c10;border:1px solid #202c38;border-radius:8px;padding:12px;white-space:pre-wrap;word-break:break-word;min-height:54px;max-height:300px;overflow:auto}
.row{display:flex;gap:7px;flex-wrap:wrap;align-items:center}.metricbar{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin:14px 0}.metric{padding:13px;background:#121c26;border:1px solid var(--line);border-radius:9px}.metric b{display:block;font-size:19px;margin-top:4px}.metric span{color:var(--muted);font-size:11px;text-transform:uppercase}
.flash{animation:flash .7s ease}@keyframes flash{0%{background:#365383}100%{background:transparent}}.event{padding:5px 0;border-bottom:1px solid #202b36;color:#9fc6ff}
.badge{display:inline-block;padding:4px 8px;border-radius:99px;background:#1a2b40;color:#a9c9ff;font-size:11px}.danger{color:#ff9c9c}.success{color:#8cebb2}
@media(max-width:900px){.layout{grid-template-columns:1fr}nav{display:flex;overflow:auto;border-right:0;border-bottom:1px solid var(--line)}.tab{width:auto;white-space:nowrap}.grid{grid-template-columns:1fr}.metricbar{grid-template-columns:repeat(2,1fr)}}
</style>

<style id="webgo-art-theme">
:root{color-scheme:dark;--wg-bg:#080808;--wg-panel:#11110f;--wg-line:#34342f;--wg-ink:#f1f1ea;--wg-muted:#aaa99e;--wg-acid:#d9ff4f;--wg-coral:#ff665a;--wg-blue:#79a7ff;--wg-cyan:#62e6dc;--wg-shadow:0 26px 80px #0009}
html,body{background:var(--wg-bg)!important;color:var(--wg-ink)!important}
body{position:relative;min-height:100vh;overflow-x:hidden}
body:before{content:"";position:fixed;z-index:-3;inset:0;background:radial-gradient(circle at 82% -5%,#4b32ff44 0 14%,transparent 31%),linear-gradient(151deg,transparent 0 38%,#ff5a4f33 38% 44%,transparent 44% 100%),linear-gradient(28deg,transparent 0 72%,#7fffe722 72% 78%,transparent 78% 100%),#080808}
body:after{content:"";position:fixed;z-index:-2;width:430px;height:430px;right:-150px;top:-170px;border:56px solid var(--wg-acid);border-radius:50%;opacity:.52;transform:rotate(22deg);pointer-events:none}
h1,h2,h3{letter-spacing:-.035em}h1{font-weight:900!important}
button,.btn{border-radius:2px!important;border:1px solid var(--wg-line)!important;background:#1a1a18!important;color:var(--wg-ink)!important;box-shadow:none!important;transition:transform .12s ease,filter .12s ease!important}
button:hover,.btn:hover{transform:translateY(-1px);filter:brightness(1.14)}
button.primary,.btn.primary{background:var(--wg-acid)!important;color:#111!important;border-color:var(--wg-acid)!important}
button.danger,.btn.danger{background:#421d20!important;color:#ffaaa4!important;border-color:#6b292c!important}
input,select,textarea,.input,.select,.textarea{border-radius:2px!important;background:#090909!important;color:white!important;border:1px solid #3d3d38!important}
.card,.panel,.modal,.column,.task{border-radius:2px!important;border-color:var(--wg-line)!important;box-shadow:var(--wg-shadow)}
pre{border-radius:2px!important;background:#080808!important;border:1px solid #30302c!important;color:#b9cae6!important}
table{background:#0e0e0e}th{color:#8f8f86!important;text-transform:uppercase;letter-spacing:.08em}
a{color:var(--wg-blue)}
</style>
</head>
<body>
<div aria-hidden="true" style="position:fixed;z-index:-1;left:-110px;top:38%;width:420px;height:78px;background:#ff665a99;transform:rotate(-31deg);pointer-events:none"></div>
<div aria-hidden="true" style="position:fixed;z-index:-1;right:24%;top:13%;width:170px;height:470px;background:linear-gradient(#62e6dc00,#62e6dc40,#62e6dc00);transform:rotate(14deg);pointer-events:none"></div>
<header><div class="brand">Web<span>Go</span> Showcase</div><div class="native" id="bootState">Booting...</div></header>
<div class="layout">
<nav>
<button class="tab active" data-view="overview">Overview</button>
<button class="tab" data-view="bindings">Bindings</button>
<button class="tab" data-view="javascript">Go → JS</button>
<button class="tab" data-view="messages">Messages</button>
<button class="tab" data-view="window">Window</button>
<button class="tab" data-view="errors">Errors</button>
<button class="tab" data-view="navigation">Navigation</button>
</nav>
<main>
<section class="view active" id="overview">
<div class="hero"><span class="badge">Comprehensive Demo</span><h1>WebGo Framework Showcase</h1><div class="muted">A single app exercising the public WebGo API and native JavaScript bridge.</div></div>
<div class="metricbar"><div class="metric"><span>WebView2</span><b id="mVersion">...</b></div><div class="metric"><span>HWND</span><b id="mHwnd">...</b></div><div class="metric"><span>Buttons</span><b id="mButtons">...</b></div><div class="metric"><span>Init Script</span><b id="mInit">...</b></div></div>
<div class="grid"><div class="panel"><h2>Runtime information</h2><p>Queries native values through a Go binding.</p><button class="btn primary" id="runtimeBtn">Refresh native info</button><pre id="runtimeOut"></pre></div><div class="panel"><h2>Background work</h2><p>The bound Go function sleeps outside the WebView2 event callback, while the frontend remains responsive.</p><input class="input" id="jobSeconds" type="number" min="1" max="8" value="3"><button class="btn green" id="jobBtn">Run job</button><pre id="jobOut">No job running.</pre></div></div>
</section>

<section class="view" id="bindings">
<div class="hero"><h1>JavaScript → Go Bindings</h1><div class="muted">Strings, structs, arrays, numbers and Promise-based return values.</div></div>
<div class="grid">
<div class="panel"><h2>Simple string binding</h2><div class="row"><input class="input" id="helloName" value="World"><button class="btn primary" id="helloBtn">Call Go</button></div><pre id="helloOut"></pre></div>
<div class="panel"><h2>Structured request/response</h2><div class="row"><input class="input" id="calcName" value="sample"><input class="input" id="calcCount" type="number" value="5"><input class="input" id="calcRate" type="number" value="2.5" step=".1"><button class="btn primary" id="calcBtn">Calculate</button></div><pre id="calcOut"></pre></div>
</div>
</section>

<section class="view" id="javascript">
<div class="hero"><h1>Go → JavaScript</h1><div class="muted">Eval and EvalWithResult both run through the WebView UI thread.</div></div>
<div class="grid"><div class="panel"><h2>Eval</h2><input class="input" id="evalText" value="native update"><button class="btn primary" id="evalBtn">Go calls Eval()</button><pre id="evalTarget">Waiting for Go.</pre></div><div class="panel"><h2>EvalWithResult</h2><button class="btn primary" id="inspectBtn">Inspect this page</button><pre id="inspectOut">No inspection yet.</pre></div></div>
</section>

<section class="view" id="messages">
<div class="hero"><h1>Raw WebMessages</h1><div class="muted">Go publishes application events independently of the Promise binding protocol.</div></div>
<div class="panel"><button class="btn primary" id="eventsBtn">Start event stream</button><button class="btn" id="clearEvents">Clear</button><pre id="eventsOut">Waiting...</pre></div>
</section>

<section class="view" id="window">
<div class="hero"><h1>Native Window Control</h1><div class="muted">Title, size hints, DevTools and native HWND access.</div></div>
<div class="grid"><div class="panel"><h2>Window title</h2><input class="input" id="titleText" value="Renamed WebGo Window"><button class="btn primary" id="renameBtn">Rename</button><pre id="windowOut"></pre></div><div class="panel"><h2>Size constraints</h2><div class="row"><button class="btn" data-size="normal">Normal</button><button class="btn" data-size="fixed">Fixed</button><button class="btn" data-size="minimum">Minimum</button><button class="btn" data-size="maximum">Maximum</button></div><button class="btn" id="devBtn">Open DevTools</button></div></div>
</section>

<section class="view" id="errors">
<div class="hero"><h1>Error and Panic Handling</h1><div class="muted">Go errors reject JavaScript Promises, and binding panics are recovered by the WebGo bridge.</div></div>
<div class="grid"><div class="panel"><h2>Return a Go error</h2><input class="input" id="errorText" value="demo error from Go"><button class="btn red" id="errorBtn">Trigger error</button><pre id="errorOut"></pre></div><div class="panel"><h2>Panic recovery</h2><button class="btn red" id="panicBtn">Trigger Go panic</button><pre id="panicOut"></pre></div></div>
</section>

<section class="view" id="navigation">
<div class="hero"><h1>Navigation Policy</h1><div class="muted">Only about:blank, data: and example.com are allowed by this demo.</div></div>
<div class="panel"><button class="btn green" id="allowedBtn">Navigate to example.com</button><button class="btn red" id="blockedBtn">Try github.com</button><pre id="navOut">The blocked navigation should leave this page in place.</pre></div>
</section>
</main>
</div>
<script>
var $=function(s){return document.querySelector(s)};
document.querySelectorAll(".tab").forEach(function(t){t.onclick=function(){document.querySelectorAll(".tab").forEach(function(x){x.classList.remove("active")});document.querySelectorAll(".view").forEach(function(x){x.classList.remove("active")});t.classList.add("active");$("#"+t.dataset.view).classList.add("active")}});
$("#bootState").textContent=window.showcaseBoot&&window.showcaseBoot.injectedByGo?"Init script active":"Init script missing";
$("#mInit").textContent=window.showcaseBoot?"YES":"NO";
async function refreshRuntime(){var x=await showRuntimeInfo();$("#runtimeOut").textContent=JSON.stringify(x,null,2);$("#mVersion").textContent=x.webview2Version;$("#mHwnd").textContent=x.hwnd;$("#mButtons").textContent=document.querySelectorAll("button").length}
$("#runtimeBtn").onclick=refreshRuntime;
$("#jobBtn").onclick=async function(){var out=$("#jobOut");out.textContent="Working...";try{out.textContent=JSON.stringify(await showBackgroundJob(Number($("#jobSeconds").value)),null,2)}catch(e){out.textContent=e.message}};
$("#helloBtn").onclick=async function(){$("#helloOut").textContent=await showHello($("#helloName").value)};
$("#calcBtn").onclick=async function(){try{$("#calcOut").textContent=JSON.stringify(await showCalculate({name:$("#calcName").value,count:Number($("#calcCount").value),rate:Number($("#calcRate").value)}),null,2)}catch(e){$("#calcOut").textContent=e.message}};
$("#evalBtn").onclick=async function(){$("#evalTarget").textContent=await showEvalFromGo($("#evalText").value)};
$("#inspectBtn").onclick=async function(){$("#inspectOut").textContent=await showInspectPage()};
$("#eventsBtn").onclick=async function(){$("#eventsOut").textContent=await showStartEvents()+"\n"};
$("#clearEvents").onclick=function(){$("#eventsOut").textContent=""};
window.chrome.webview.addEventListener("message",function(e){var out=$("#eventsOut");out.textContent+=String(e.data)+"\n";out.scrollTop=out.scrollHeight});
$("#renameBtn").onclick=async function(){$("#windowOut").textContent=await showRename($("#titleText").value)};
document.querySelectorAll("[data-size]").forEach(function(b){b.onclick=async function(){$("#windowOut").textContent=await showResize(b.dataset.size)}});
$("#devBtn").onclick=async function(){$("#windowOut").textContent=await showOpenDevTools()};
$("#errorBtn").onclick=async function(){try{await showForceError($("#errorText").value);$("#errorOut").textContent="Unexpected success"}catch(e){$("#errorOut").textContent="Rejected Promise: "+e.message}};
$("#panicBtn").onclick=async function(){try{await showPanicRecovery();$("#panicOut").textContent="Unexpected success"}catch(e){$("#panicOut").textContent="Recovered panic: "+e.message}};
$("#allowedBtn").onclick=async function(){$("#navOut").textContent=await showNavigateAllowed()};
$("#blockedBtn").onclick=async function(){$("#navOut").textContent=await showNavigateBlocked()};
refreshRuntime();
</script>
</body>
</html>`