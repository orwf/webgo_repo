//go:build windows && amd64

// Package webgo is a lightweight native Windows desktop framework built on
// Microsoft Edge WebView2.
//
// It combines a Go backend with an HTML/CSS/JavaScript frontend and provides
// Promise-based JavaScript-to-Go bindings, JavaScript execution from Go,
// native Win32 window access, navigation filtering, per-window dispatching,
// WebView2 DevTools, and direct HTML loading.
//
// Basic usage:
//
//	package main
//
//	import webgo "github.com/orwf/webgo_repo"
//
//	func main() {
//		w := webgo.New(true)
//		defer w.Destroy()
//
//		w.SetTitle("My WebGo App")
//		w.SetSize(1000, 700, webgo.HintNone)
//
//		w.NavigateToString(`
//			<!doctype html>
//			<html>
//			<body>
//				<h1>Hello from WebGo</h1>
//			</body>
//			</html>
//		`)
//
//		w.Run()
//	}
package webgo
