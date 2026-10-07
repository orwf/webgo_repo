//go:build windows && amd64

package main

import (
	"errors"
	"log"
	"strings"

	"webgo_repo-main/webview"
)

type UserInput struct {
	Name  string \`json:"name"\`
	Email string \`json:"email"\`
	Age   int    \`json:"age"\`
}

type UserResult struct {
	Valid   bool     \`json:"valid"\`
	Summary string   \`json:"summary"\`
	Tags    []string \`json:"tags"\`
}

func main() {
	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("WebGo Demo 03 - Structs and Errors")
	w.SetSize(1000, 720, webview.HintNone)

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

	w.NavigateToString(\`<!doctype html>
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
</head>
<body>
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
</html>\`)

	w.Run()
}
