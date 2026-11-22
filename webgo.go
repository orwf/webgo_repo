//go:build windows && amd64

// Package webgo provides a lightweight native Windows WebView2 framework for Go.
//
// The package exposes the public WebGo API from the module root so applications
// can simply import:
//
//	import webgo "github.com/orwf/webgo_repo"
//
// WebGo uses Microsoft Edge WebView2 for rendering and Win32/COM for native
// integration. Application UIs can be written with HTML, CSS, and JavaScript
// while native logic remains in Go.
package webgo

import "github.com/orwf/webgo_repo/webview"

// Hint controls window resize constraints.
type Hint = webview.Hint

const (
	// HintNone changes the window size without adding a resize constraint.
	HintNone = webview.HintNone

	// HintFixed locks the window to the supplied dimensions.
	HintFixed = webview.HintFixed

	// HintMin sets the minimum window dimensions.
	HintMin = webview.HintMin

	// HintMax sets the maximum window dimensions.
	HintMax = webview.HintMax
)

// WindowOptions configures the native application window.
type WindowOptions = webview.WindowOptions

// WebViewOptions configures WebGo and WebView2 creation.
type WebViewOptions = webview.WebViewOptions

// DownloadInfo describes a WebView2 download starting in the page.
type DownloadInfo = webview.DownloadInfo

// DownloadDecision configures how WebView2 starts a download.
type DownloadDecision = webview.DownloadDecision

// PermissionKind identifies the capability a page is requesting.
type PermissionKind = webview.PermissionKind

// PermissionState is the app's decision for a permission request.
type PermissionState = webview.PermissionState

// PermissionRequest contains the origin and context of a permission request.
type PermissionRequest = webview.PermissionRequest

const (
	PermissionKindUnknown                     = webview.PermissionKindUnknown
	PermissionKindMicrophone                  = webview.PermissionKindMicrophone
	PermissionKindCamera                      = webview.PermissionKindCamera
	PermissionKindGeolocation                 = webview.PermissionKindGeolocation
	PermissionKindNotifications               = webview.PermissionKindNotifications
	PermissionKindOtherSensors                = webview.PermissionKindOtherSensors
	PermissionKindClipboardRead               = webview.PermissionKindClipboardRead
	PermissionKindMultipleAutomaticDownloads  = webview.PermissionKindMultipleAutomaticDownloads
	PermissionKindFileReadWrite               = webview.PermissionKindFileReadWrite
	PermissionKindAutoplay                    = webview.PermissionKindAutoplay
	PermissionKindLocalFonts                  = webview.PermissionKindLocalFonts
	PermissionKindMidiSystemExclusiveMessages = webview.PermissionKindMidiSystemExclusiveMessages
	PermissionKindWindowManagement            = webview.PermissionKindWindowManagement
	PermissionStateDefault                    = webview.PermissionStateDefault
	PermissionStateAllow                      = webview.PermissionStateAllow
	PermissionStateDeny                       = webview.PermissionStateDeny
)

// WebView is the main WebGo application interface.
//
// It provides:
//   - native window lifecycle
//   - navigation
//   - HTML loading
//   - JavaScript execution
//   - Go <-> JavaScript bindings
//   - navigation filtering
//   - native WebView messages
//   - DevTools access
//   - native HWND access
type WebView = webview.WebView

// New creates a WebView using default options.
//
// Debug controls whether WebView2 developer tools and default browser
// context menus are enabled.
func New(debug bool) WebView {
	return webview.New(debug)
}

// NewWithOptions creates a WebView using custom options.
//
// This convenience constructor panics if creation fails.
// Production applications should generally prefer NewWithOptionsE.
func NewWithOptions(opts WebViewOptions) WebView {
	return webview.NewWithOptions(opts)
}

// NewWithOptionsE creates a WebView using custom options and returns
// creation errors to the caller.
func NewWithOptionsE(opts WebViewOptions) (WebView, error) {
	return webview.NewWithOptionsE(opts)
}
