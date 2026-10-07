//go:build windows && amd64

package main

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	webgo "github.com/orwf/webgo_repo"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	Owner       string `json:"owner"`
	Due         string `json:"due"`
	Created     string `json:"created"`
}

type TaskSummary struct {
	Total      int `json:"total"`
	Backlog    int `json:"backlog"`
	InProgress int `json:"inProgress"`
	Done       int `json:"done"`
	High       int `json:"high"`
}

type TaskStore struct {
	mu     sync.RWMutex
	tasks  []Task
	nextID int
}

func newTaskStore() *TaskStore {
	now := time.Now()
	return &TaskStore{
		nextID: 9,
		tasks: []Task{
			{1, "Implement WebView2 recovery", "Add ProcessFailed handling and restore state.", "done", "high", "Alex", now.AddDate(0,0,-1).Format("2006-01-02"), now.Add(-72*time.Hour).Format("2006-01-02")},
			{2, "Build import facade", "Expose the package from the module root.", "done", "high", "Sam", now.Format("2006-01-02"), now.Add(-48*time.Hour).Format("2006-01-02")},
			{3, "Add comprehensive demos", "Create larger real-world examples.", "in-progress", "high", "Alex", now.AddDate(0,0,2).Format("2006-01-02"), now.Add(-24*time.Hour).Format("2006-01-02")},
			{4, "Improve API docs", "Document edge cases and recovery behavior.", "backlog", "medium", "Chris", now.AddDate(0,0,5).Format("2006-01-02"), now.Add(-24*time.Hour).Format("2006-01-02")},
			{5, "Add CI workflow", "Build and test Windows amd64 automatically.", "backlog", "medium", "Sam", now.AddDate(0,0,7).Format("2006-01-02"), now.Format("2006-01-02")},
			{6, "Virtual host mapping", "Serve bundled frontend assets cleanly.", "backlog", "low", "Chris", now.AddDate(0,0,12).Format("2006-01-02"), now.Format("2006-01-02")},
			{7, "Permission event API", "Expose camera, mic, clipboard permissions.", "backlog", "medium", "Alex", now.AddDate(0,0,14).Format("2006-01-02"), now.Format("2006-01-02")},
			{8, "Release v0.1.0", "Tag a stable importable release.", "in-progress", "high", "Sam", now.AddDate(0,0,3).Format("2006-01-02"), now.Format("2006-01-02")},
		},
	}
}

func (s *TaskStore) List(query, status, priority string) []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if query != "" &&
			!strings.Contains(strings.ToLower(t.Title), query) &&
			!strings.Contains(strings.ToLower(t.Description), query) &&
			!strings.Contains(strings.ToLower(t.Owner), query) {
			continue
		}
		if status != "" && status != "all" && t.Status != status {
			continue
		}
		if priority != "" && priority != "all" && t.Priority != priority {
			continue
		}
		out = append(out, t)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (s *TaskStore) Summary() TaskSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out TaskSummary
	out.Total = len(s.tasks)
	for _, t := range s.tasks {
		switch t.Status {
		case "backlog":
			out.Backlog++
		case "in-progress":
			out.InProgress++
		case "done":
			out.Done++
		}
		if t.Priority == "high" {
			out.High++
		}
	}
	return out
}

func (s *TaskStore) Create(title, description, priority, owner, due string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, errors.New("task title is required")
	}
	if priority != "low" && priority != "medium" && priority != "high" {
		return Task{}, errors.New("invalid priority")
	}
	if strings.TrimSpace(owner) == "" {
		owner = "Unassigned"
	}

	t := Task{
		ID:          s.nextID,
		Title:       title,
		Description: strings.TrimSpace(description),
		Status:      "backlog",
		Priority:    priority,
		Owner:       strings.TrimSpace(owner),
		Due:         due,
		Created:     time.Now().Format("2006-01-02"),
	}
	s.nextID++
	s.tasks = append(s.tasks, t)
	return t, nil
}

func (s *TaskStore) UpdateStatus(id int, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if status != "backlog" && status != "in-progress" && status != "done" {
		return errors.New("invalid task status")
	}

	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Status = status
			return nil
		}
	}
	return errors.New("task not found")
}

func (s *TaskStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}
	return errors.New("task not found")
}

func main() {
	store := newTaskStore()

	w, err := webgo.NewWithOptionsE(webgo.WebViewOptions{
		Debug: true,
		AutoFocus: true,
		Window: webgo.WindowOptions{
			Title: "WebGo Comprehensive Demo - Project Manager",
			Width: 1440,
			Height: 900,
		},
	})
	if err != nil { log.Fatal(err) }
	defer w.Destroy()

	mustBind := func(name string, fn interface{}) {
		if err := w.Bind(name, fn); err != nil { log.Fatal(err) }
	}

	mustBind("taskList", store.List)
	mustBind("taskSummary", store.Summary)
	mustBind("taskCreate", store.Create)
	mustBind("taskUpdateStatus", store.UpdateStatus)
	mustBind("taskDelete", store.Delete)
	mustBind("taskAppInfo", func() map[string]string {
		return map[string]string{
			"webview2": w.Version(),
			"hwnd": fmt.Sprintf("0x%X", w.HWND()),
			"started": time.Now().Format("15:04:05"),
		}
	})
	mustBind("taskOpenDevTools", func() string {
		w.OpenDevTools()
		return "Developer tools opened"
	})

	w.NavigateToString(taskManagerHTML)
	w.Run()
}

const taskManagerHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>WebGo Project Manager</title>
<style>
:root{color-scheme:dark;--bg:#0b0f14;--panel:#111820;--panel2:#161f29;--line:#293543;--ink:#eef3f8;--muted:#8898aa;--blue:#5b8cff;--green:#4bd58a;--amber:#f3b956;--red:#ef6a6a}
*{box-sizing:border-box}body{margin:0;background:#0b0f14;color:var(--ink);font:14px system-ui,Segoe UI,sans-serif}.shell{display:grid;grid-template-columns:250px 1fr;min-height:100vh}
aside{background:#0e141b;border-right:1px solid var(--line);padding:22px}.brand{font-weight:800;font-size:22px;margin-bottom:26px}.brand span{color:var(--blue)}.nav a{display:block;color:var(--muted);padding:10px 12px;margin:4px 0;border-radius:8px;text-decoration:none}.nav a.active{background:#172230;color:white}
main{padding:24px;min-width:0}.top{display:flex;justify-content:space-between;align-items:center;gap:16px}.top h1{margin:0;font-size:28px}.muted{color:var(--muted)}.toolbar{display:flex;gap:8px;flex-wrap:wrap}
button,input,select,textarea{font:inherit}.btn{border:1px solid var(--line);background:#17212b;color:white;border-radius:8px;padding:9px 13px;cursor:pointer}.btn.primary{background:var(--blue);border-color:transparent}.btn.danger{background:#452126;color:#ffb3b3}.small{font-size:12px;padding:5px 8px}
.cards{display:grid;grid-template-columns:repeat(5,1fr);gap:12px;margin:20px 0}.card{background:var(--panel);border:1px solid var(--line);border-radius:11px;padding:15px}.metric{font-size:27px;font-weight:800;margin-top:6px}.label{font-size:11px;color:var(--muted);text-transform:uppercase;letter-spacing:.08em}
.filters{display:flex;gap:8px;margin:12px 0 16px;flex-wrap:wrap}.input,.select,.textarea{background:#090e13;border:1px solid var(--line);color:white;padding:8px 10px;border-radius:7px;outline:none}.input{min-width:260px}
.board{display:grid;grid-template-columns:repeat(3,minmax(280px,1fr));gap:14px}.column{background:#0e151d;border:1px solid var(--line);border-radius:11px;min-height:440px;overflow:hidden}.column-head{padding:13px 14px;border-bottom:1px solid var(--line);font-weight:700;display:flex;justify-content:space-between}.column-body{padding:10px;display:grid;gap:9px}
.task{background:#151e27;border:1px solid #2a3847;border-radius:9px;padding:12px}.task h3{font-size:14px;margin:0 0 7px}.task p{color:var(--muted);margin:0 0 10px;line-height:1.4}.meta{display:flex;gap:8px;flex-wrap:wrap;font-size:11px;color:#9fb0c2}.tag{padding:3px 7px;border-radius:99px;background:#202b37}.high{background:#49252b;color:#ff9f9f}.medium{background:#46391f;color:#ffd17b}.low{background:#173829;color:#89eab4}.task-actions{display:flex;gap:5px;margin-top:10px}
.modal-bg{display:none;position:fixed;inset:0;background:#000a;align-items:center;justify-content:center}.modal-bg.show{display:flex}.modal{width:520px;max-width:calc(100vw - 30px);background:#111820;border:1px solid var(--line);border-radius:12px;padding:20px}.field{margin:11px 0}.field label{display:block;color:var(--muted);margin-bottom:5px}.field input,.field select,.field textarea{width:100%}.field textarea{height:90px;resize:vertical}
.toast{display:none;position:fixed;right:20px;bottom:20px;background:#15202b;border:1px solid var(--line);padding:11px 14px;border-radius:9px}.toast.show{display:block}
@media(max-width:1050px){.shell{grid-template-columns:1fr}aside{display:none}.board{grid-template-columns:1fr}.cards{grid-template-columns:repeat(2,1fr)}}
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
<div class="shell">
<aside><div class="brand">Web<span>Go</span> PM</div><div class="nav"><a class="active">Board</a><a>Backlog</a><a>Milestones</a><a>People</a><a>Reports</a></div></aside>
<main>
<div class="top"><div><h1>Project Board</h1><div class="muted">State and business logic are handled by Go</div></div><div class="toolbar"><button class="btn" id="infoBtn">Runtime</button><button class="btn" id="devBtn">DevTools</button><button class="btn primary" id="newBtn">New Task</button></div></div>
<div class="cards"><div class="card"><div class="label">Total</div><div class="metric" id="sTotal">0</div></div><div class="card"><div class="label">Backlog</div><div class="metric" id="sBacklog">0</div></div><div class="card"><div class="label">In Progress</div><div class="metric" id="sProgress">0</div></div><div class="card"><div class="label">Done</div><div class="metric" id="sDone">0</div></div><div class="card"><div class="label">High Priority</div><div class="metric" id="sHigh">0</div></div></div>
<div class="filters"><input class="input" id="query" placeholder="Search title, description or owner"><select class="select" id="priority"><option value="all">All priorities</option><option>high</option><option>medium</option><option>low</option></select></div>
<div class="board">
<div class="column"><div class="column-head"><span>Backlog</span><span id="cBacklog">0</span></div><div class="column-body" id="backlog"></div></div>
<div class="column"><div class="column-head"><span>In Progress</span><span id="cProgress">0</span></div><div class="column-body" id="progress"></div></div>
<div class="column"><div class="column-head"><span>Done</span><span id="cDone">0</span></div><div class="column-body" id="done"></div></div>
</div>
</main>
</div>
<div class="modal-bg" id="modal"><div class="modal"><h2>Create Task</h2><div class="field"><label>Title</label><input class="input" id="title"></div><div class="field"><label>Description</label><textarea class="textarea" id="desc"></textarea></div><div class="field"><label>Owner</label><input class="input" id="owner" value="Alex"></div><div class="field"><label>Priority</label><select class="select" id="newPriority"><option>medium</option><option>high</option><option>low</option></select></div><div class="field"><label>Due date</label><input class="input" id="due" type="date"></div><div class="toolbar"><button class="btn" id="cancelBtn">Cancel</button><button class="btn primary" id="createBtn">Create</button></div></div></div>
<div class="toast" id="toast"></div>
<script>
var $=function(s){return document.querySelector(s)},toastTimer=0;
function esc(v){var d=document.createElement("div");d.textContent=String(v);return d.innerHTML}
function toast(m){var e=$("#toast");e.textContent=m;e.classList.add("show");clearTimeout(toastTimer);toastTimer=setTimeout(function(){e.classList.remove("show")},2200)}
async function summary(){var s=await taskSummary();$("#sTotal").textContent=s.total;$("#sBacklog").textContent=s.backlog;$("#sProgress").textContent=s.inProgress;$("#sDone").textContent=s.done;$("#sHigh").textContent=s.high}
function card(t){
  var h="<div class='task'><h3>"+esc(t.title)+"</h3><p>"+esc(t.description||"No description")+"</p><div class='meta'><span class='tag "+t.priority+"'>"+t.priority+"</span><span class='tag'>"+esc(t.owner)+"</span><span class='tag'>due "+esc(t.due||"none")+"</span></div><div class='task-actions'>";
  if(t.status!=="backlog")h+="<button class='btn small' data-move='"+t.id+"' data-to='backlog'>Backlog</button>";
  if(t.status!=="in-progress")h+="<button class='btn small' data-move='"+t.id+"' data-to='in-progress'>Start</button>";
  if(t.status!=="done")h+="<button class='btn small' data-move='"+t.id+"' data-to='done'>Done</button>";
  h+="<button class='btn small danger' data-delete='"+t.id+"'>Delete</button></div></div>";
  return h;
}
async function loadBoard(){
  var q=$("#query").value,p=$("#priority").value;
  var all=await taskList(q,"all",p),b=[],i=[],d=[];
  all.forEach(function(t){if(t.status==="backlog")b.push(t);else if(t.status==="in-progress")i.push(t);else d.push(t)});
  $("#backlog").innerHTML=b.map(card).join("");$("#progress").innerHTML=i.map(card).join("");$("#done").innerHTML=d.map(card).join("");
  $("#cBacklog").textContent=b.length;$("#cProgress").textContent=i.length;$("#cDone").textContent=d.length;
  document.querySelectorAll("[data-move]").forEach(function(x){x.onclick=async function(){try{await taskUpdateStatus(Number(x.dataset.move),x.dataset.to);await refresh();toast("Task moved")}catch(e){toast(e.message)}}});
  document.querySelectorAll("[data-delete]").forEach(function(x){x.onclick=async function(){if(!confirm("Delete this task?"))return;try{await taskDelete(Number(x.dataset.delete));await refresh();toast("Task deleted")}catch(e){toast(e.message)}}});
}
async function refresh(){await Promise.all([summary(),loadBoard()])}
$("#query").oninput=loadBoard;$("#priority").onchange=loadBoard;$("#newBtn").onclick=function(){$("#modal").classList.add("show")};$("#cancelBtn").onclick=function(){$("#modal").classList.remove("show")};
$("#createBtn").onclick=async function(){try{await taskCreate($("#title").value,$("#desc").value,$("#newPriority").value,$("#owner").value,$("#due").value);$("#title").value="";$("#desc").value="";$("#modal").classList.remove("show");await refresh();toast("Task created")}catch(e){toast(e.message)}};
$("#infoBtn").onclick=async function(){alert(JSON.stringify(await taskAppInfo(),null,2))};$("#devBtn").onclick=async function(){toast(await taskOpenDevTools())};
refresh();
</script>
</body>
</html>`