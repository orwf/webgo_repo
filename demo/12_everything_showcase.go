//go:build windows && amd64

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"sync/atomic"
	"time"

	webgo "github.com/orwf/webgo_repo"
)

// This helper intentionally never runs. It keeps minimal compile-time examples of
// the two convenience constructors next to the production-style NewWithOptionsE
// constructor used by main.
//
// Creating multiple WebViews casually from arbitrary goroutines is intentionally
// avoided here because WebView2 is STA/thread-affine.
func constructorExamples() {
	if false {
		a := webgo.New(true)
		a.Destroy()

		b := webgo.NewWithOptions(webgo.WebViewOptions{
			Debug:     true,
			AutoFocus: true,
			Window: webgo.WindowOptions{
				Title:  "Constructor Example",
				Width:  900,
				Height: 600,
			},
		})
		b.Destroy()
	}
}

type ArtPayload struct {
	Name      string    `json:"name"`
	Intensity float64   `json:"intensity"`
	Count     int       `json:"count"`
	Tags      []string  `json:"tags"`
	Points    []float64 `json:"points"`
}

type ArtResult struct {
	Title    string    `json:"title"`
	Spectrum []float64 `json:"spectrum"`
	Average  float64   `json:"average"`
	Message  string    `json:"message"`
}

type RuntimeSnapshot struct {
	WebView2 string `json:"webview2"`
	HWND     string `json:"hwnd"`
	Time     string `json:"time"`
	Uptime   string `json:"uptime"`
	Messages int64  `json:"messages"`
}

func main() {
	constructorExamples()

	started := time.Now()
	var messageCount int64

	w, err := webgo.NewWithOptionsE(webgo.WebViewOptions{
		Debug:     true,
		AutoFocus: true,
		Window: webgo.WindowOptions{
			Title:  "WebGo API Museum",
			Width:  1480,
			Height: 920,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer w.Destroy()

	// Init() runs before each future document.
	w.Init(`
window.__webgoMuseumInit = {
  active: true,
  injectedAt: new Date().toISOString(),
  note: "Installed by WebGo.Init before document creation"
};
`)

	// SetNavigationHandler demonstrates synchronous native navigation policy.
	w.SetNavigationHandler(func(uri string) bool {
		log.Println("[navigation]", uri)

		switch {
		case uri == "about:blank":
			return true
		case strings.HasPrefix(uri, "data:"):
			return true
		case strings.HasPrefix(uri, "https://example.com"):
			return true
		default:
			log.Println("[navigation] blocked:", uri)
			return false
		}
	})

	mustBind := func(name string, fn interface{}) {
		if err := w.Bind(name, fn); err != nil {
			log.Fatalf("Bind(%s): %v", name, err)
		}
	}

	// Bind(): primitive values.
	mustBind("apiHello", func(name string) string {
		name = strings.TrimSpace(name)
		if name == "" {
			name = "visitor"
		}
		return "Hello " + name + " — returned by a native Go binding."
	})

	// Bind(): structured object in + structured object out.
	mustBind("apiGenerateArtData", func(in ArtPayload) (ArtResult, error) {
		if strings.TrimSpace(in.Name) == "" {
			return ArtResult{}, errors.New("name is required")
		}
		if in.Count < 1 || in.Count > 32 {
			return ArtResult{}, errors.New("count must be between 1 and 32")
		}
		if math.IsNaN(in.Intensity) || math.IsInf(in.Intensity, 0) {
			return ArtResult{}, errors.New("intensity must be finite")
		}

		spectrum := make([]float64, 0, in.Count)
		var total float64
		for i := 0; i < in.Count; i++ {
			v := math.Sin(float64(i)*0.62)*in.Intensity +
				math.Cos(float64(i)*0.27)*(in.Intensity*0.45)
			spectrum = append(spectrum, math.Round(v*1000)/1000)
			total += v
		}

		return ArtResult{
			Title:    "Spectrum for " + in.Name,
			Spectrum: spectrum,
			Average:  math.Round((total/float64(in.Count))*1000) / 1000,
			Message:  fmt.Sprintf("Go generated %d spectrum points from %d tags", in.Count, len(in.Tags)),
		}, nil
	})

	// Bind(): multiple return values become an array in JS.
	mustBind("apiMultipleReturns", func() (int, string, bool) {
		return 42, "multiple Go return values", true
	})

	// Bind(): returned error becomes a rejected JS Promise.
	mustBind("apiReturnError", func(message string) error {
		message = strings.TrimSpace(message)
		if message == "" {
			message = "intentional demonstration error"
		}
		return errors.New(message)
	})

	// Bind(): WebGo catches a panic inside a bound function.
	mustBind("apiPanic", func() string {
		panic("intentional demo panic — recovered by WebGo callBinding")
	})

	// Version() + HWND().
	mustBind("apiRuntime", func() RuntimeSnapshot {
		return RuntimeSnapshot{
			WebView2: w.Version(),
			HWND:     fmt.Sprintf("0x%X", w.HWND()),
			Time:     time.Now().Format(time.RFC3339),
			Uptime:   time.Since(started).Round(time.Millisecond).String(),
			Messages: atomic.LoadInt64(&messageCount),
		}
	})

	// SetTitle().
	mustBind("apiSetTitle", func(title string) string {
		title = strings.TrimSpace(title)
		if title == "" {
			title = "WebGo API Museum"
		}
		w.SetTitle(title)
		return "SetTitle applied to the native HWND."
	})

	// SetSize() with every Hint value.
	mustBind("apiSetSize", func(mode string) string {
		switch mode {
		case "none":
			w.SetSize(1260, 800, webgo.HintNone)
			return "HintNone: 1260×800, freely resizable"
		case "fixed":
			w.SetSize(1040, 720, webgo.HintFixed)
			return "HintFixed: locked to 1040×720"
		case "min":
			w.SetSize(760, 560, webgo.HintMin)
			return "HintMin: minimum 760×560"
		case "max":
			w.SetSize(1560, 980, webgo.HintMax)
			return "HintMax: maximum 1560×980"
		default:
			return "unknown sizing mode"
		}
	})

	// Eval().
	mustBind("apiEval", func(text string) string {
		encoded, _ := json.Marshal("Eval() says: " + text)
		w.Eval(fmt.Sprintf(
			`document.querySelector("#evalCanvas").textContent=%s; document.querySelector("#evalCanvas").classList.add("pulse"); setTimeout(function(){document.querySelector("#evalCanvas").classList.remove("pulse")},900);`,
			encoded,
		))
		return "Eval queued through WebGo.Dispatch."
	})

	// EvalDirect() must only run when already on the WebView UI thread, so this
	// binding explicitly combines Dispatch() + EvalDirect().
	mustBind("apiEvalDirect", func(text string) string {
		encoded, _ := json.Marshal("EvalDirect() on UI thread: " + text)
		w.Dispatch(func() {
			w.EvalDirect(fmt.Sprintf(
				`document.querySelector("#directCanvas").textContent=%s; document.querySelector("#directCanvas").classList.add("pulse"); setTimeout(function(){document.querySelector("#directCanvas").classList.remove("pulse")},900);`,
				encoded,
			))
		})
		return "Dispatch + EvalDirect scheduled."
	})

	// Dispatch() directly.
	mustBind("apiDispatch", func() string {
		w.Dispatch(func() {
			now, _ := json.Marshal(time.Now().Format("15:04:05.000"))
			w.EvalDirect(fmt.Sprintf(
				`document.querySelector("#dispatchOut").textContent="Dispatch executed on UI thread at " + %s;`,
				now,
			))
		})
		return "Native function queued with Dispatch."
	})

	// EvalWithResult().
	mustBind("apiInspectDOM", func() string {
		w.EvalWithResult(`({
  title: document.title,
  href: location.href,
  viewport: [innerWidth, innerHeight],
  elements: document.querySelectorAll("*").length,
  buttons: document.querySelectorAll("button").length,
  activeSection: document.querySelector(".section.active") && document.querySelector(".section.active").id,
  initObject: window.__webgoMuseumInit
})`, func(result string) {
			encoded, _ := json.Marshal(result)
			w.Eval(fmt.Sprintf(
				`document.querySelector("#inspectOut").textContent=%s;`,
				encoded,
			))
		})
		return "EvalWithResult started asynchronously."
	})

	// PostMessage().
	mustBind("apiStartMessageStream", func() string {
		go func() {
			for i := 1; i <= 12; i++ {
				atomic.AddInt64(&messageCount, 1)
				payload, _ := json.Marshal(map[string]interface{}{
					"type":  "museum-event",
					"index": i,
					"time":  time.Now().Format("15:04:05.000"),
					"value": math.Round(math.Sin(float64(i)*0.7)*1000) / 1000,
				})
				w.PostMessage(string(payload))
				time.Sleep(450 * time.Millisecond)
			}
		}()
		return "12 raw WebMessages started."
	})

	// OpenDevTools().
	mustBind("apiDevTools", func() string {
		w.OpenDevTools()
		return "OpenDevTools requested."
	})

	// Long-running binding work executes outside the WebView2 callback.
	mustBind("apiBackgroundWork", func(milliseconds int) (map[string]interface{}, error) {
		if milliseconds < 100 || milliseconds > 6000 {
			return nil, errors.New("milliseconds must be between 100 and 6000")
		}

		begin := time.Now()
		time.Sleep(time.Duration(milliseconds) * time.Millisecond)

		return map[string]interface{}{
			"requestedMs": milliseconds,
			"actualMs":    time.Since(begin).Milliseconds(),
			"completed":   time.Now().Format(time.RFC3339Nano),
		}, nil
	})

	// Navigate(). The app returns to NavigateToString() automatically.
	mustBind("apiNavigateAllowed", func() string {
		go func() {
			time.Sleep(2200 * time.Millisecond)
			w.NavigateToString(allFeaturesHTML)
		}()
		w.Navigate("https://example.com")
		return "Navigate(example.com) requested; museum will restore in ~2.2 seconds."
	})

	// Blocked Navigate().
	mustBind("apiNavigateBlocked", func() string {
		w.Navigate("https://github.com")
		return "Navigate(github.com) requested; SetNavigationHandler should cancel it."
	})

	// NavigateToString().
	mustBind("apiReloadMuseumHTML", func() string {
		go func() {
			time.Sleep(150 * time.Millisecond)
			w.NavigateToString(allFeaturesHTML)
		}()
		return "NavigateToString will reload the complete museum document."
	})

	// Temporary binding used to demonstrate Unbind().
	mustBind("temporaryBinding", func() string {
		return "temporaryBinding is alive."
	})
	mustBind("apiUnbindTemporary", func() string {
		w.Unbind("temporaryBinding")
		return "Unbind removed temporaryBinding from the Go registry and current page."
	})

	// Terminate(). This closes the native HWND and Run() exits.
	mustBind("apiTerminate", func() string {
		go func() {
			time.Sleep(250 * time.Millisecond)
			w.Terminate()
		}()
		return "Terminate posted WM_CLOSE. The application will close."
	})

	// NavigateToString() is the initial page load.
	w.NavigateToString(allFeaturesHTML)

	// Run() owns the Win32 message loop until the window closes.
	w.Run()

	// Destroy() executes through defer above after Run() returns.
}

const allFeaturesHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>WebGo API Museum</title>
<style>
:root{
 color-scheme:dark;
 --ink:#f1f1ea;--muted:#aaa99e;--black:#080808;--paper:#151515;--paper2:#1d1d1b;
 --line:#34342f;--acid:#d9ff4f;--coral:#ff665a;--blue:#79a7ff;--lav:#b69cff;--cyan:#62e6dc;
 --shadow:0 26px 80px #0009;
}
*{box-sizing:border-box}
html,body{margin:0;min-height:100%;background:var(--black);color:var(--ink);font:14px Inter,Segoe UI,system-ui,sans-serif}
body{overflow-x:hidden}
.art-bg{position:fixed;inset:0;pointer-events:none;overflow:hidden;z-index:0}
.shape{position:absolute;filter:blur(.1px);opacity:.7;mix-blend-mode:screen}
.s1{width:420px;height:420px;border:55px solid #5d43ff;border-radius:50%;right:-130px;top:-160px;transform:rotate(23deg)}
.s2{width:480px;height:90px;background:#ff4b3f;left:-160px;top:42%;transform:rotate(-31deg)}
.s3{width:300px;height:300px;border:1px solid #ffffff33;left:34%;bottom:-210px;transform:rotate(45deg)}
.s4{width:200px;height:500px;background:linear-gradient(#77ffe400, #77ffe455, #77ffe400);right:23%;top:15%;transform:rotate(15deg)}
.noise{position:fixed;inset:0;opacity:.035;pointer-events:none;background-image:url("data:image/svg+xml,%3Csvg viewBox='0 0 180 180' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.8' numOctaves='5' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='.8'/%3E%3C/svg%3E")}
.shell{position:relative;z-index:1;display:grid;grid-template-columns:265px minmax(0,1fr);min-height:100vh}
aside{position:sticky;top:0;height:100vh;padding:26px 19px;border-right:1px solid var(--line);background:#0c0c0ceb;backdrop-filter:blur(18px)}
.logo{font-size:27px;font-weight:900;letter-spacing:-.06em;margin-bottom:4px}.logo i{font-style:normal;color:var(--acid)}
.kicker{color:var(--muted);font-size:11px;text-transform:uppercase;letter-spacing:.16em;margin-bottom:28px}
.nav{display:grid;gap:5px}.nav button{border:0;background:transparent;color:#99998f;text-align:left;padding:10px 11px;border-radius:4px;cursor:pointer;font:inherit}.nav button:hover,.nav button.active{background:var(--ink);color:#0a0a0a}.nav .num{display:inline-block;width:27px;color:#62625c}.nav button.active .num{color:#5a5a52}
.side-foot{position:absolute;bottom:20px;left:19px;right:19px;border-top:1px solid var(--line);padding-top:13px;color:#77776f;font-size:11px;line-height:1.55}
main{padding:26px 28px 70px;min-width:0}.section{display:none}.section.active{display:block}
.hero{min-height:310px;display:grid;grid-template-columns:1.3fr .7fr;gap:16px;margin-bottom:18px}
.hero-copy{position:relative;overflow:hidden;padding:30px;background:#eeeeea;color:#111;border-radius:2px;box-shadow:var(--shadow)}
.hero-copy:after{content:"API";position:absolute;right:-24px;bottom:-75px;font-size:190px;font-weight:950;letter-spacing:-.12em;color:#1111110d}
.eyebrow{text-transform:uppercase;letter-spacing:.16em;font-size:11px;font-weight:800}.hero h1{font-size:64px;line-height:.86;letter-spacing:-.075em;margin:54px 0 17px;max-width:730px}.hero p{font-size:15px;max-width:620px;line-height:1.55;color:#444}
.hero-art{position:relative;background:#3c2cff;overflow:hidden;min-height:300px}
.hero-art .circle{position:absolute;width:250px;height:250px;border:48px solid var(--acid);border-radius:50%;top:-42px;right:-45px}
.hero-art .bar{position:absolute;background:var(--coral);height:75px;width:420px;bottom:65px;left:-100px;transform:rotate(-25deg)}
.hero-art .type{position:absolute;left:25px;bottom:20px;font-size:11px;letter-spacing:.16em;text-transform:uppercase}
.metrics{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin:12px 0 18px}.metric{background:#101010;border:1px solid var(--line);padding:14px}.metric span{display:block;color:var(--muted);font-size:10px;text-transform:uppercase;letter-spacing:.12em}.metric b{display:block;margin-top:7px;font-size:17px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px}.grid.three{grid-template-columns:repeat(3,minmax(0,1fr))}
.card{position:relative;background:#111;border:1px solid var(--line);padding:17px;min-width:0}.card.acid{background:var(--acid);color:#101010}.card.coral{background:var(--coral);color:#141414}.card.blue{background:#4b6dff;color:white}.card h2{font-size:16px;margin:0 0 6px}.card p{color:var(--muted);line-height:1.45;margin:0 0 13px}.card.acid p,.card.coral p{color:#24241f}.api{font:11px Consolas,monospace;color:var(--cyan);margin-bottom:10px}.card.acid .api,.card.coral .api{color:#38382f}
.row{display:flex;gap:7px;align-items:center;flex-wrap:wrap}.stack{display:grid;gap:8px}
button,input,select{font:inherit}.btn{cursor:pointer;border:1px solid #45453f;background:#1a1a18;color:#eee;padding:8px 11px;border-radius:2px}.btn:hover{background:#252522}.btn.light{background:#efefe9;color:#111;border-color:#efefe9}.btn.danger{background:#3a1718;color:#ffaaa4;border-color:#6b292c}.btn.acid{background:var(--acid);color:#111;border-color:var(--acid)}
input,select{background:#090909;border:1px solid #3d3d38;color:white;padding:8px 9px;border-radius:2px;outline:none;min-width:150px}.card.acid input,.card.coral input{background:#ffffff99;color:#111;border-color:#1114}
pre,.output{margin-top:10px;background:#080808;border:1px solid #30302c;padding:10px;min-height:48px;max-height:260px;overflow:auto;white-space:pre-wrap;word-break:break-word;font:12px Consolas,monospace;color:#b9cae6}.card.acid pre,.card.coral pre{background:#111;color:#e8ead9}
.section-head{display:flex;justify-content:space-between;align-items:end;gap:20px;margin:8px 0 18px}.section-head h1{font-size:42px;letter-spacing:-.055em;margin:0}.section-head p{color:var(--muted);max-width:500px;text-align:right}
.api-table{width:100%;border-collapse:collapse;background:#0e0e0e;border:1px solid var(--line)}.api-table th,.api-table td{padding:10px;border-bottom:1px solid #2b2b27;text-align:left}.api-table th{color:#8f8f86;font-size:10px;text-transform:uppercase;letter-spacing:.1em}.yes{color:#9cffbd}.note{color:#dbd575}
.spectrum{display:flex;height:120px;align-items:center;gap:3px;margin-top:11px}.spectrum i{display:block;flex:1;background:var(--lav);min-width:2px}
.message-line{padding:6px 0;border-bottom:1px solid #252522;color:#9bc5ff}
.pulse{animation:pulse .9s ease}@keyframes pulse{0%{box-shadow:inset 0 0 0 999px #d9ff4f;color:#111}100%{box-shadow:inset 0 0 0 999px transparent}}
.warning{padding:12px;border-left:3px solid var(--coral);background:#ff665a12;color:#ffc0bb}
.constructor{font:12px Consolas,monospace;white-space:pre-wrap;background:#080808;padding:14px;border:1px solid #2b2b27;color:#a8c3f0}
@media(max-width:1100px){.shell{grid-template-columns:1fr}aside{position:relative;height:auto;border-right:0;border-bottom:1px solid var(--line)}.nav{grid-template-columns:repeat(4,1fr)}.side-foot{display:none}.hero{grid-template-columns:1fr}.hero h1{font-size:50px}.grid,.grid.three{grid-template-columns:1fr}.metrics{grid-template-columns:repeat(2,1fr)}}
</style>
</head>
<body>
<div class="art-bg"><div class="shape s1"></div><div class="shape s2"></div><div class="shape s3"></div><div class="shape s4"></div></div><div class="noise"></div>
<div class="shell">
<aside>
<div class="logo">WEB<i>GO</i></div><div class="kicker">API Museum / Windows</div>
<div class="nav">
<button class="active" data-section="home"><span class="num">01</span> Overview</button>
<button data-section="bindings"><span class="num">02</span> Bindings</button>
<button data-section="javascript"><span class="num">03</span> JavaScript</button>
<button data-section="window"><span class="num">04</span> Window</button>
<button data-section="messages"><span class="num">05</span> Messages</button>
<button data-section="navigation"><span class="num">06</span> Navigation</button>
<button data-section="lifecycle"><span class="num">07</span> Lifecycle</button>
<button data-section="reference"><span class="num">08</span> API Map</button>
</div>
<div class="side-foot">Microsoft Edge WebView2<br>Win32 + COM<br>Go backend / HTML frontend</div>
</aside>

<main>
<section class="section active" id="home">
<div class="hero">
<div class="hero-copy"><div class="eyebrow">Flagship demonstration</div><h1>EVERY PUBLIC API.<br>ONE WINDOW.</h1><p>This museum is a self-contained WebGo application designed to exercise the complete public WebView interface while presenting it as a modern-art desktop experience.</p></div>
<div class="hero-art"><div class="circle"></div><div class="bar"></div><div class="type">Go ↔ JavaScript<br>WebView2 / STA UI</div></div>
</div>
<div class="metrics"><div class="metric"><span>WebView2</span><b id="metricVersion">loading</b></div><div class="metric"><span>HWND</span><b id="metricHwnd">loading</b></div><div class="metric"><span>Init()</span><b id="metricInit">loading</b></div><div class="metric"><span>Raw messages</span><b id="metricMessages">0</b></div></div>
<div class="grid">
<div class="card"><div class="api">Version() · HWND()</div><h2>Native runtime snapshot</h2><p>Reads the live WebView2 runtime version and native Win32 window handle through Go.</p><button class="btn acid" id="runtimeBtn">Refresh Runtime</button><pre id="runtimeOut"></pre></div>
<div class="card"><div class="api">Bind() · goroutine-safe async work</div><h2>Background native work</h2><p>The binding sleeps in Go while the UI remains responsive because WebGo does not execute arbitrary bound functions inside the WebView2 message callback.</p><div class="row"><input id="workMs" type="number" value="1800"><button class="btn" id="workBtn">Run Native Job</button></div><pre id="workOut">Ready.</pre></div>
</div>
</section>

<section class="section" id="bindings">
<div class="section-head"><h1>Bindings</h1><p>Promise-based JavaScript calls into reflected Go functions, including structs, slices, multiple returns, errors and panic recovery.</p></div>
<div class="grid">
<div class="card acid"><div class="api">Bind("apiHello", func(string) string)</div><h2>Primitive arguments</h2><div class="row"><input id="helloName" value="Museum Visitor"><button class="btn light" id="helloBtn">Call Go</button></div><pre id="helloOut"></pre></div>
<div class="card"><div class="api">Struct → Go → Struct</div><h2>Structured JSON binding</h2><div class="row"><input id="artName" value="Signal No. 7"><input id="artIntensity" type="number" step=".1" value="2.4"><input id="artCount" type="number" value="14"><button class="btn acid" id="artBtn">Generate</button></div><pre id="artOut"></pre><div class="spectrum" id="spectrum"></div></div>
<div class="card"><div class="api">Multiple returns</div><h2>Tuple-style return</h2><button class="btn" id="multiBtn">Return (int, string, bool)</button><pre id="multiOut"></pre></div>
<div class="card coral"><div class="api">error · panic recovery</div><h2>Failure propagation</h2><div class="row"><button class="btn light" id="errorBtn">Return Go Error</button><button class="btn danger" id="panicBtn">Panic in Go</button></div><pre id="failureOut"></pre></div>
<div class="card"><div class="api">Unbind()</div><h2>Dynamic bridge removal</h2><div class="row"><button class="btn" id="tempCallBtn">Call temporaryBinding</button><button class="btn danger" id="unbindBtn">Unbind It</button></div><pre id="unbindOut"></pre></div>
</div>
</section>

<section class="section" id="javascript">
<div class="section-head"><h1>JavaScript Execution</h1><p>Three paths: safe dispatched Eval, direct UI-thread EvalDirect, and asynchronous EvalWithResult.</p></div>
<div class="grid three">
<div class="card blue"><div class="api">Eval()</div><h2>Dispatched execution</h2><input id="evalText" value="hello from Go"><button class="btn light" id="evalBtn">Execute</button><pre id="evalCanvas">Waiting.</pre></div>
<div class="card"><div class="api">Dispatch() + EvalDirect()</div><h2>Direct UI-thread execution</h2><input id="directText" value="direct native update"><button class="btn acid" id="directBtn">Dispatch Direct</button><pre id="directCanvas">Waiting.</pre></div>
<div class="card"><div class="api">EvalWithResult()</div><h2>Read browser state</h2><button class="btn" id="inspectBtn">Inspect DOM</button><pre id="inspectOut">No result yet.</pre></div>
</div>
<div class="card" style="margin-top:12px"><div class="api">Dispatch()</div><h2>Explicit UI dispatch</h2><button class="btn acid" id="dispatchBtn">Queue Native UI Work</button><pre id="dispatchOut">Not dispatched yet.</pre></div>
</section>

<section class="section" id="window">
<div class="section-head"><h1>Native Window</h1><p>WebGo keeps a native HWND while WebView2 renders the frontend.</p></div>
<div class="grid">
<div class="card"><div class="api">SetTitle()</div><h2>Window title</h2><div class="row"><input id="titleInput" value="WebGo — Renamed Museum"><button class="btn acid" id="titleBtn">Set Native Title</button></div><pre id="windowOut"></pre></div>
<div class="card"><div class="api">SetSize() · HintNone · HintFixed · HintMin · HintMax</div><h2>Sizing constraints</h2><div class="row"><button class="btn" data-size="none">None</button><button class="btn" data-size="fixed">Fixed</button><button class="btn" data-size="min">Minimum</button><button class="btn" data-size="max">Maximum</button></div></div>
<div class="card acid"><div class="api">OpenDevTools()</div><h2>Chromium Developer Tools</h2><p>Available because this demo starts with Debug enabled.</p><button class="btn light" id="devtoolsBtn">Open DevTools</button></div>
</div>
</section>

<section class="section" id="messages">
<div class="section-head"><h1>Raw Messages</h1><p>PostMessage operates independently from the Bind RPC protocol and is useful for event streams.</p></div>
<div class="card"><div class="api">PostMessage()</div><div class="row"><button class="btn acid" id="streamBtn">Start 12 Messages</button><button class="btn" id="clearMessages">Clear</button></div><pre id="messageOut">No messages yet.</pre></div>
</section>

<section class="section" id="navigation">
<div class="section-head"><h1>Navigation</h1><p>Navigate, NavigateToString and SetNavigationHandler all participate in the document lifecycle.</p></div>
<div class="grid">
<div class="card"><div class="api">Navigate()</div><h2>Allowed remote navigation</h2><p>Loads example.com, then Go automatically restores this museum using NavigateToString.</p><button class="btn acid" id="allowedNav">Visit example.com</button><pre id="navOut"></pre></div>
<div class="card coral"><div class="api">SetNavigationHandler()</div><h2>Blocked origin</h2><p>github.com is intentionally blocked by the native policy.</p><button class="btn light" id="blockedNav">Attempt github.com</button></div>
<div class="card"><div class="api">NavigateToString()</div><h2>Recreate this whole document</h2><button class="btn" id="reloadHtml">Reload Museum HTML</button></div>
</div>
</section>

<section class="section" id="lifecycle">
<div class="section-head"><h1>Lifecycle</h1><p>Run owns the message loop. Terminate closes the window. Destroy releases native resources after Run returns.</p></div>
<div class="grid">
<div class="card"><div class="api">Run()</div><h2>Message loop</h2><p class="yes">ACTIVE NOW</p><p>The program is currently inside WebGo's Win32 message loop.</p></div>
<div class="card"><div class="api">Destroy()</div><h2>Deferred cleanup</h2><p>main uses <strong>defer w.Destroy()</strong>, so cleanup runs immediately after Run exits.</p></div>
<div class="card coral"><div class="api">Terminate()</div><h2>Close this demo</h2><p>This intentionally posts WM_CLOSE and ends the application.</p><button class="btn light" id="terminateBtn">Terminate Application</button></div>
</div>
<div class="warning" style="margin-top:13px">The constructors New() and NewWithOptions() are included as compile-time examples in constructorExamples(). main uses NewWithOptionsE(), which is the recommended production constructor because it returns startup errors.</div>
</section>

<section class="section" id="reference">
<div class="section-head"><h1>Complete API Map</h1><p>The current public WebView interface and root-package constructors represented in this demo.</p></div>
<table class="api-table">
<thead><tr><th>API</th><th>Used</th><th>Where</th></tr></thead>
<tbody>
<tr><td>New(debug)</td><td class="yes">✓ compile example</td><td>constructorExamples</td></tr>
<tr><td>NewWithOptions(opts)</td><td class="yes">✓ compile example</td><td>constructorExamples</td></tr>
<tr><td>NewWithOptionsE(opts)</td><td class="yes">✓ live</td><td>main</td></tr>
<tr><td>Run()</td><td class="yes">✓ live</td><td>application message loop</td></tr>
<tr><td>Terminate()</td><td class="yes">✓ live</td><td>Lifecycle section</td></tr>
<tr><td>Dispatch(fn)</td><td class="yes">✓ live</td><td>JavaScript section</td></tr>
<tr><td>Destroy()</td><td class="yes">✓ live</td><td>deferred cleanup</td></tr>
<tr><td>SetTitle()</td><td class="yes">✓ live</td><td>Window section</td></tr>
<tr><td>SetSize()</td><td class="yes">✓ live</td><td>all four Hint values</td></tr>
<tr><td>Navigate()</td><td class="yes">✓ live</td><td>Navigation section</td></tr>
<tr><td>NavigateToString()</td><td class="yes">✓ live</td><td>initial + reload</td></tr>
<tr><td>Init()</td><td class="yes">✓ live</td><td>window.__webgoMuseumInit</td></tr>
<tr><td>Eval()</td><td class="yes">✓ live</td><td>JavaScript section</td></tr>
<tr><td>EvalWithResult()</td><td class="yes">✓ live</td><td>DOM inspector</td></tr>
<tr><td>EvalDirect()</td><td class="yes">✓ live</td><td>inside Dispatch</td></tr>
<tr><td>Bind()</td><td class="yes">✓ live</td><td>all controls</td></tr>
<tr><td>Unbind()</td><td class="yes">✓ live</td><td>temporaryBinding</td></tr>
<tr><td>PostMessage()</td><td class="yes">✓ live</td><td>message stream</td></tr>
<tr><td>OpenDevTools()</td><td class="yes">✓ live</td><td>Window section</td></tr>
<tr><td>Version()</td><td class="yes">✓ live</td><td>runtime snapshot</td></tr>
<tr><td>HWND()</td><td class="yes">✓ live</td><td>runtime snapshot</td></tr>
<tr><td>SetNavigationHandler()</td><td class="yes">✓ live</td><td>origin allow/block policy</td></tr>
</tbody></table>

<h2 style="margin-top:24px">Constructor reference</h2>
<div class="constructor">webgo.New(true)

webgo.NewWithOptions(webgo.WebViewOptions{
    Debug: true,
    AutoFocus: true,
    Window: webgo.WindowOptions{
        Title: "Example",
        Width: 900,
        Height: 600,
    },
})

webgo.NewWithOptionsE(...) // used by this running application</div>
</section>
</main>
</div>

<script>
var $=function(s){return document.querySelector(s)};
document.querySelectorAll(".nav button").forEach(function(btn){
 btn.onclick=function(){
  document.querySelectorAll(".nav button").forEach(function(x){x.classList.remove("active")});
  document.querySelectorAll(".section").forEach(function(x){x.classList.remove("active")});
  btn.classList.add("active");
  $("#"+btn.dataset.section).classList.add("active");
 };
});

$("#metricInit").textContent=window.__webgoMuseumInit&&window.__webgoMuseumInit.active?"ACTIVE":"MISSING";

async function refreshRuntime(){
 var r=await apiRuntime();
 $("#metricVersion").textContent=r.webview2;
 $("#metricHwnd").textContent=r.hwnd;
 $("#metricMessages").textContent=r.messages;
 $("#runtimeOut").textContent=JSON.stringify(r,null,2);
}
$("#runtimeBtn").onclick=refreshRuntime;

$("#workBtn").onclick=async function(){
 var o=$("#workOut");o.textContent="Go is working — UI should remain interactive...";
 try{o.textContent=JSON.stringify(await apiBackgroundWork(Number($("#workMs").value)),null,2)}
 catch(e){o.textContent=e.message}
};

$("#helloBtn").onclick=async function(){$("#helloOut").textContent=await apiHello($("#helloName").value)};

$("#artBtn").onclick=async function(){
 try{
  var r=await apiGenerateArtData({name:$("#artName").value,intensity:Number($("#artIntensity").value),count:Number($("#artCount").value),tags:["modern","native","webview"],points:[1,2,3]});
  $("#artOut").textContent=JSON.stringify(r,null,2);
  var max=0;r.spectrum.forEach(function(v){max=Math.max(max,Math.abs(v))});if(max===0)max=1;
  var html="";r.spectrum.forEach(function(v){var h=15+Math.round(Math.abs(v)/max*100);html+="<i style='height:"+h+"px;opacity:"+(v>=0?1:.42)+"'></i>"});
  $("#spectrum").innerHTML=html;
 }catch(e){$("#artOut").textContent=e.message}
};

$("#multiBtn").onclick=async function(){$("#multiOut").textContent=JSON.stringify(await apiMultipleReturns(),null,2)};
$("#errorBtn").onclick=async function(){try{await apiReturnError("A deliberate Go error crossed the bridge");$("#failureOut").textContent="unexpected success"}catch(e){$("#failureOut").textContent="Rejected Promise: "+e.message}};
$("#panicBtn").onclick=async function(){try{await apiPanic();$("#failureOut").textContent="unexpected success"}catch(e){$("#failureOut").textContent="Recovered panic: "+e.message}};

$("#tempCallBtn").onclick=async function(){try{$("#unbindOut").textContent=await temporaryBinding()}catch(e){$("#unbindOut").textContent="Unavailable: "+e.message}};
$("#unbindBtn").onclick=async function(){$("#unbindOut").textContent=await apiUnbindTemporary()};

$("#evalBtn").onclick=async function(){$("#evalCanvas").textContent=await apiEval($("#evalText").value)};
$("#directBtn").onclick=async function(){$("#directCanvas").textContent=await apiEvalDirect($("#directText").value)};
$("#inspectBtn").onclick=async function(){$("#inspectOut").textContent=await apiInspectDOM()};
$("#dispatchBtn").onclick=async function(){$("#dispatchOut").textContent=await apiDispatch()};

$("#titleBtn").onclick=async function(){$("#windowOut").textContent=await apiSetTitle($("#titleInput").value)};
document.querySelectorAll("[data-size]").forEach(function(btn){btn.onclick=async function(){$("#windowOut").textContent=await apiSetSize(btn.dataset.size)}});
$("#devtoolsBtn").onclick=async function(){$("#windowOut").textContent=await apiDevTools()};

$("#streamBtn").onclick=async function(){$("#messageOut").textContent=(await apiStartMessageStream())+"\n"};
$("#clearMessages").onclick=function(){$("#messageOut").textContent=""};
window.chrome.webview.addEventListener("message",function(e){
 var o=$("#messageOut");
 o.textContent+=String(e.data)+"\n";
 o.scrollTop=o.scrollHeight;
 var n=Number($("#metricMessages").textContent)||0;
 $("#metricMessages").textContent=n+1;
});

$("#allowedNav").onclick=async function(){$("#navOut").textContent=await apiNavigateAllowed()};
$("#blockedNav").onclick=async function(){$("#navOut").textContent=await apiNavigateBlocked()};
$("#reloadHtml").onclick=async function(){$("#navOut").textContent=await apiReloadMuseumHTML()};

$("#terminateBtn").onclick=async function(){if(confirm("Terminate the native WebGo application?"))await apiTerminate()};

refreshRuntime();
</script>
</body>
</html>`