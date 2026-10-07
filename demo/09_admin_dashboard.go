//go:build windows && amd64

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	webgo "github.com/orwf/webgo_repo"
)

type DemoUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	LastSeen string `json:"lastSeen"`
}

type DemoStats struct {
	Users     int `json:"users"`
	Online    int `json:"online"`
	Admins    int `json:"admins"`
	Suspended int `json:"suspended"`
}

type DemoEvent struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type AdminDemoStore struct {
	mu     sync.RWMutex
	users  []DemoUser
	events []DemoEvent
	nextID int
}

func newAdminDemoStore() *AdminDemoStore {
	return &AdminDemoStore{
		nextID: 9,
		users: []DemoUser{
			{1, "alice", "admin", "online", "now"},
			{2, "bob", "user", "offline", "12 min ago"},
			{3, "charlie", "moderator", "online", "now"},
			{4, "diana", "user", "suspended", "2 hours ago"},
			{5, "eve", "user", "offline", "1 day ago"},
			{6, "frank", "user", "online", "now"},
			{7, "grace", "moderator", "offline", "45 min ago"},
			{8, "henry", "user", "online", "now"},
		},
		events: []DemoEvent{
			{time.Now().Add(-2 * time.Minute).Format("15:04:05"), "info", "Admin dashboard started"},
			{time.Now().Add(-70 * time.Second).Format("15:04:05"), "success", "Native Go bindings registered"},
			{time.Now().Add(-20 * time.Second).Format("15:04:05"), "warning", "Demo warning event"},
		},
	}
}

func (s *AdminDemoStore) addEvent(level, message string) {
	s.events = append([]DemoEvent{{time.Now().Format("15:04:05"), level, message}}, s.events...)
	if len(s.events) > 80 {
		s.events = s.events[:80]
	}
}

func (s *AdminDemoStore) listUsers(query, status string) []DemoUser {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	status = strings.ToLower(strings.TrimSpace(status))
	out := make([]DemoUser, 0, len(s.users))

	for _, u := range s.users {
		if query != "" &&
			!strings.Contains(strings.ToLower(u.Username), query) &&
			!strings.Contains(strings.ToLower(u.Role), query) {
			continue
		}
		if status != "" && status != "all" && u.Status != status {
			continue
		}
		out = append(out, u)
	}
	return out
}

func (s *AdminDemoStore) stats() DemoStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st := DemoStats{Users: len(s.users)}
	for _, u := range s.users {
		if u.Status == "online" { st.Online++ }
		if u.Status == "suspended" { st.Suspended++ }
		if u.Role == "admin" { st.Admins++ }
	}
	return st
}

func (s *AdminDemoStore) getEvents() []DemoEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]DemoEvent(nil), s.events...)
}

func (s *AdminDemoStore) createUser(username, role string) (DemoUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	username = strings.TrimSpace(username)
	role = strings.ToLower(strings.TrimSpace(role))
	if username == "" {
		return DemoUser{}, errors.New("username is required")
	}
	if role != "user" && role != "moderator" && role != "admin" {
		return DemoUser{}, errors.New("invalid role")
	}
	for _, u := range s.users {
		if strings.EqualFold(u.Username, username) {
			return DemoUser{}, errors.New("username already exists")
		}
	}

	u := DemoUser{ID: s.nextID, Username: username, Role: role, Status: "offline", LastSeen: "never"}
	s.nextID++
	s.users = append(s.users, u)
	s.addEvent("success", "Created user "+username)
	return u, nil
}

func (s *AdminDemoStore) setStatus(id int, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if status != "online" && status != "offline" && status != "suspended" {
		return errors.New("invalid status")
	}
	for i := range s.users {
		if s.users[i].ID == id {
			s.users[i].Status = status
			s.users[i].LastSeen = "just now"
			s.addEvent("info", fmt.Sprintf("Changed %s to %s", s.users[i].Username, status))
			return nil
		}
	}
	return errors.New("user not found")
}

func (s *AdminDemoStore) deleteUser(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, u := range s.users {
		if u.ID == id {
			s.users = append(s.users[:i], s.users[i+1:]...)
			s.addEvent("warning", "Deleted user "+u.Username)
			return nil
		}
	}
	return errors.New("user not found")
}

func main() {
	store := newAdminDemoStore()

	w, err := webgo.NewWithOptionsE(webgo.WebViewOptions{
		Debug: true,
		AutoFocus: true,
		Window: webgo.WindowOptions{
			Title: "WebGo Comprehensive Demo - Admin Dashboard",
			Width: 1380,
			Height: 860,
		},
	})
	if err != nil { log.Fatal(err) }
	defer w.Destroy()

	mustBind := func(name string, fn interface{}) {
		if err := w.Bind(name, fn); err != nil { log.Fatal(err) }
	}

	mustBind("demoListUsers", store.listUsers)
	mustBind("demoStats", store.stats)
	mustBind("demoEvents", store.getEvents)
	mustBind("demoCreateUser", store.createUser)
	mustBind("demoSetStatus", store.setStatus)
	mustBind("demoDeleteUser", store.deleteUser)
	mustBind("demoNativeInfo", func() map[string]string {
		return map[string]string{
			"webview2": w.Version(),
			"hwnd": fmt.Sprintf("0x%X", w.HWND()),
			"time": time.Now().Format(time.RFC1123),
		}
	})
	mustBind("demoOpenDevTools", func() string {
		w.OpenDevTools()
		return "DevTools opened"
	})

	w.NavigateToString(adminDemoHTML)
	w.Run()
}

const adminDemoHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>WebGo Admin Dashboard</title>
<style>
:root{color-scheme:dark;--bg:#091018;--panel:#101923;--panel2:#162330;--line:#283849;--ink:#edf3f8;--muted:#8da0b5;--blue:#4b8cff;--green:#55d894;--amber:#ffc45b;--red:#ff7373}
*{box-sizing:border-box}body{margin:0;background:radial-gradient(circle at 80% 0,#16283b,#091018 42%);color:var(--ink);font:14px system-ui,Segoe UI,sans-serif}.app{display:grid;grid-template-columns:230px 1fr;min-height:100vh}
aside{padding:24px 18px;border-right:1px solid var(--line);background:#0b121bcc}.brand{font-size:22px;font-weight:800;margin-bottom:28px}.brand b{color:var(--blue)}.nav{display:grid;gap:5px}.nav div{padding:10px 12px;border-radius:8px;color:var(--muted)}.nav .active{background:#172432;color:white}
main{padding:26px;min-width:0}.top{display:flex;justify-content:space-between;align-items:center;gap:20px;margin-bottom:20px}.top h1{margin:0;font-size:27px}.sub{color:var(--muted);margin-top:4px}.toolbar{display:flex;gap:8px;flex-wrap:wrap}
button,input,select{font:inherit}.btn{cursor:pointer;border:1px solid var(--line);background:#172330;color:white;padding:9px 13px;border-radius:8px}.btn:hover{filter:brightness(1.15)}.btn.primary{background:var(--blue);border-color:transparent}.btn.danger{background:#401f25;color:#ffb8b8}.small{padding:5px 8px;font-size:12px}
.cards{display:grid;grid-template-columns:repeat(4,1fr);gap:13px;margin-bottom:16px}.card,.panel{background:linear-gradient(180deg,#111c27,#0e171f);border:1px solid var(--line);border-radius:12px}.card{padding:17px}.metric{font-size:30px;font-weight:800;margin-top:7px}.label{text-transform:uppercase;letter-spacing:.08em;font-size:11px;color:var(--muted)}
.grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(280px,1fr);gap:15px}.head{padding:14px 15px;border-bottom:1px solid var(--line);display:flex;justify-content:space-between;gap:10px;align-items:center}.body{padding:14px}.filters{display:flex;gap:7px}
input,select{background:#091018;border:1px solid var(--line);color:white;padding:8px 10px;border-radius:7px;outline:none}table{width:100%;border-collapse:collapse}th,td{padding:10px;border-bottom:1px solid #202e3b;text-align:left}th{font-size:11px;color:var(--muted);text-transform:uppercase}.role{color:#a7c8ff}
.pill{padding:4px 8px;border-radius:99px;font-size:11px}.online{background:#173a2a;color:#86efb2}.offline{background:#252e39;color:#b4c0cb}.suspended{background:#44242a;color:#ffa6a6}.actions{display:flex;gap:5px}
.events{max-height:500px;overflow:auto}.event{padding:10px 2px;border-bottom:1px solid #202e3b}.event-time{font-size:11px;color:var(--muted)}.success{color:#83eeb0}.warning{color:#ffd47a}
.modal-bg{display:none;position:fixed;inset:0;background:#000a;align-items:center;justify-content:center}.modal-bg.show{display:flex}.modal{width:420px;max-width:calc(100vw - 30px);padding:20px;background:#111b25;border:1px solid var(--line);border-radius:12px}.field{margin:12px 0}.field label{display:block;color:var(--muted);margin-bottom:5px}.field input,.field select{width:100%}
.toast{display:none;position:fixed;right:20px;bottom:20px;padding:11px 14px;border:1px solid var(--line);background:#14202b;border-radius:9px}.toast.show{display:block}
@media(max-width:960px){.app{grid-template-columns:1fr}aside{display:none}.grid{grid-template-columns:1fr}.cards{grid-template-columns:repeat(2,1fr)}}
</style>
</head>
<body>
<div class="app">
<aside><div class="brand">Web<b>Go</b> Admin</div><div class="nav"><div class="active">Overview</div><div>Users</div><div>Activity</div><div>Runtime</div><div>Settings</div></div></aside>
<main>
<div class="top"><div><h1>Administration Dashboard</h1><div class="sub">Interactive application state lives in Go</div></div><div class="toolbar"><button class="btn" id="runtime">Runtime</button><button class="btn" id="devtools">DevTools</button><button class="btn primary" id="newUser">New User</button></div></div>
<div class="cards"><div class="card"><div class="label">Users</div><div class="metric" id="usersCount">0</div></div><div class="card"><div class="label">Online</div><div class="metric" id="onlineCount">0</div></div><div class="card"><div class="label">Admins</div><div class="metric" id="adminCount">0</div></div><div class="card"><div class="label">Suspended</div><div class="metric" id="suspendedCount">0</div></div></div>
<div class="grid">
<div class="panel"><div class="head"><strong>User Directory</strong><div class="filters"><input id="query" placeholder="Search users"><select id="status"><option value="all">All</option><option>online</option><option>offline</option><option>suspended</option></select></div></div><div class="body"><table><thead><tr><th>ID</th><th>User</th><th>Role</th><th>Status</th><th>Last Seen</th><th>Actions</th></tr></thead><tbody id="userRows"></tbody></table></div></div>
<div class="panel"><div class="head"><strong>Go Event Log</strong><button class="btn small" id="refreshEvents">Refresh</button></div><div class="body events" id="eventRows"></div></div>
</div>
</main>
</div>
<div class="modal-bg" id="modal"><div class="modal"><h2>Create User</h2><div class="field"><label>Username</label><input id="name"></div><div class="field"><label>Role</label><select id="role"><option>user</option><option>moderator</option><option>admin</option></select></div><div class="toolbar"><button class="btn" id="cancel">Cancel</button><button class="btn primary" id="create">Create</button></div></div></div>
<div class="toast" id="toast"></div>
<script>
var q=function(s){return document.querySelector(s)};
var toastTimer=0;
function safe(v){var d=document.createElement("div");d.textContent=String(v);return d.innerHTML}
function toast(msg){var el=q("#toast");el.textContent=msg;el.classList.add("show");clearTimeout(toastTimer);toastTimer=setTimeout(function(){el.classList.remove("show")},2200)}
async function loadStats(){var s=await demoStats();q("#usersCount").textContent=s.users;q("#onlineCount").textContent=s.online;q("#adminCount").textContent=s.admins;q("#suspendedCount").textContent=s.suspended}
async function loadUsers(){
  var rows=await demoListUsers(q("#query").value,q("#status").value);var html="";
  rows.forEach(function(u){
    html+="<tr><td>"+u.id+"</td><td><strong>"+safe(u.username)+"</strong></td><td class='role'>"+safe(u.role)+"</td><td><span class='pill "+u.status+"'>"+u.status+"</span></td><td>"+safe(u.lastSeen)+"</td><td><div class='actions'><button class='btn small' data-cycle='"+u.id+"' data-status='"+u.status+"'>Cycle</button><button class='btn small danger' data-delete='"+u.id+"'>Delete</button></div></td></tr>";
  });
  q("#userRows").innerHTML=html;
  document.querySelectorAll("[data-cycle]").forEach(function(b){b.onclick=async function(){var cur=b.dataset.status;var next=cur==="online"?"offline":cur==="offline"?"suspended":"online";try{await demoSetStatus(Number(b.dataset.cycle),next);await refreshAll();toast("Status updated")}catch(e){toast(e.message)}}});
  document.querySelectorAll("[data-delete]").forEach(function(b){b.onclick=async function(){if(!confirm("Delete this user?"))return;try{await demoDeleteUser(Number(b.dataset.delete));await refreshAll();toast("User deleted")}catch(e){toast(e.message)}}});
}
async function loadEvents(){var ev=await demoEvents();var html="";ev.forEach(function(e){html+="<div class='event'><div class='event-time'>"+safe(e.time)+" · "+safe(e.level)+"</div><div class='"+safe(e.level)+"'>"+safe(e.message)+"</div></div>"});q("#eventRows").innerHTML=html}
async function refreshAll(){await Promise.all([loadStats(),loadUsers(),loadEvents()])}
q("#query").oninput=loadUsers;q("#status").onchange=loadUsers;q("#refreshEvents").onclick=loadEvents;
q("#newUser").onclick=function(){q("#modal").classList.add("show")};q("#cancel").onclick=function(){q("#modal").classList.remove("show")};
q("#create").onclick=async function(){try{await demoCreateUser(q("#name").value,q("#role").value);q("#name").value="";q("#modal").classList.remove("show");await refreshAll();toast("User created")}catch(e){toast(e.message)}};
q("#runtime").onclick=async function(){alert(JSON.stringify(await demoNativeInfo(),null,2))};
q("#devtools").onclick=async function(){toast(await demoOpenDevTools())};
refreshAll();setInterval(loadStats,5000);
</script>
</body>
</html>`