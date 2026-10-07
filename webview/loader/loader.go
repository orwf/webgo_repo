//go:build windows && amd64

// Package loader handles loading WebView2Loader.dll.
//
// Load order:
//  1. WebView2Loader.dll already in PATH / next to exe (system install)
//  2. Embedded bytes (if you embed the DLL with //go:embed)
//  3. Error — ask user to install WebView2 runtime
package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// ─────────────────────────────────────────────────────────────────
// LAZY DLL REFERENCE
// ─────────────────────────────────────────────────────────────────

var (
	dll = loadWebView2Loader()

	procCreate = dll.NewProc("CreateCoreWebView2EnvironmentWithOptions")

	procCompareBrowserVer = dll.NewProc("CompareBrowserVersions")

	procGetBrowserVersionStr = dll.NewProc("GetAvailableCoreWebView2BrowserVersionString")
)

func loadWebView2Loader() *syscall.LazyDLL {
	var candidates []string

	// 1. Current working directory.
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "webview", "loader", "WebView2Loader.dll"),
			filepath.Join(wd, "loader", "WebView2Loader.dll"),
			filepath.Join(wd, "WebView2Loader.dll"),
		)
	}

	// 2. Directory containing the compiled executable.
	if exe, err := os.Executable(); err == nil {
		base := filepath.Dir(exe)

		candidates = append(candidates,
			filepath.Join(base, "WebView2Loader.dll"),
			filepath.Join(base, "webview", "loader", "WebView2Loader.dll"),
			filepath.Join(base, "loader", "WebView2Loader.dll"),
		)
	}

	for _, candidate := range candidates {
		abs, _ := filepath.Abs(candidate)

		fmt.Println("[WebView2] checking:", abs)

		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			fmt.Println("[WebView2] loader found:", abs)
			return syscall.NewLazyDLL(abs)
		}
	}

	panic(
		"WebView2Loader.dll not found.\n" +
			"Searched:\n  " +
			strings.Join(candidates, "\n  "),
	)
}

// ─────────────────────────────────────────────────────────────────
// CreateCoreWebView2EnvironmentWithOptions
//
// This is the entry point to WebView2. It starts the Edge browser
// process and returns an ICoreWebView2Environment via the callback.
//
// Parameters:
//   browserExecutableFolder — path to browser binary (nil = installed Edge)
//   userDataFolder          — path for user profile data (nil = default)
//   environmentOptions      — ICoreWebView2EnvironmentOptions (0 = defaults)
//   completedHandler        — ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandler
// ─────────────────────────────────────────────────────────────────

func CreateEnvironmentWithOptions(
	browserExecutableFolder *uint16,
	userDataFolder *uint16,
	environmentOptions uintptr,
	completedHandler uintptr,
) (uintptr, error) {
	if err := dll.Load(); err != nil {
		return 0, fmt.Errorf(
			"webview2amd64: failed to load WebView2Loader.dll: %w\n"+
				"  → Install the WebView2 Runtime from https://developer.microsoft.com/en-us/microsoft-edge/webview2/",
			err,
		)
	}

	r, _, err := procCreate.Call(
		uintptr(unsafe.Pointer(browserExecutableFolder)),
		uintptr(unsafe.Pointer(userDataFolder)),
		environmentOptions,
		completedHandler,
	)

	// err from SyscallN is always non-nil (it's the last errno), check HR instead
	if int32(r) < 0 {
		return 0, fmt.Errorf("CreateCoreWebView2EnvironmentWithOptions: HRESULT 0x%08X (%w)", uint32(r), err)
	}
	return r, nil
}

// ─────────────────────────────────────────────────────────────────
// GetInstalledVersion
//
// Returns the version string of the installed WebView2 runtime.
// Returns "" with no error if WebView2 is not installed.
// ─────────────────────────────────────────────────────────────────

func GetInstalledVersion() (string, error) {
	if err := dll.Load(); err != nil {
		return "", nil // not installed
	}

	var versionPtr *uint16
	r, _, _ := procGetBrowserVersionStr.Call(
		0, // nil = use default installed browser
		uintptr(unsafe.Pointer(&versionPtr)),
	)

	// HR 0x80070002 = ERROR_FILE_NOT_FOUND = not installed
	if uint32(r) == 0x80070002 {
		return "", nil
	}
	if int32(r) < 0 {
		return "", fmt.Errorf("GetAvailableCoreWebView2BrowserVersionString: HRESULT 0x%08X", uint32(r))
	}
	if versionPtr == nil {
		return "", nil
	}

	version := utf16PtrToString(versionPtr)
	coTaskMemFree(uintptr(unsafe.Pointer(versionPtr)))
	return version, nil
}

// ─────────────────────────────────────────────────────────────────
// CompareBrowserVersions
//
// Compares two version strings.
// Returns: -1 (v1 < v2), 0 (equal), 1 (v1 > v2)
// ─────────────────────────────────────────────────────────────────

func CompareBrowserVersions(v1, v2 string) (int, error) {
	if err := dll.Load(); err != nil {
		return 0, fmt.Errorf("WebView2Loader.dll not available: %w", err)
	}
	p1, _ := syscall.UTF16PtrFromString(v1)
	p2, _ := syscall.UTF16PtrFromString(v2)

	var result int32
	r, _, _ := procCompareBrowserVer.Call(
		uintptr(unsafe.Pointer(p1)),
		uintptr(unsafe.Pointer(p2)),
		uintptr(unsafe.Pointer(&result)),
	)
	if int32(r) < 0 {
		return 0, fmt.Errorf("CompareBrowserVersions: HRESULT 0x%08X", uint32(r))
	}
	return int(result), nil
}

// ─────────────────────────────────────────────────────────────────
// INTERNAL HELPERS
// ─────────────────────────────────────────────────────────────────

var ole32 = syscall.NewLazyDLL("ole32.dll")
var procFree = ole32.NewProc("CoTaskMemFree")

func coTaskMemFree(p uintptr) {
	procFree.Call(p)
}

func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	// Walk the *uint16 to find the null terminator
	n := 0
	for ptr := unsafe.Pointer(p); *(*uint16)(ptr) != 0; ptr = unsafe.Pointer(uintptr(ptr) + 2) {
		n++
	}
	slice := (*[1 << 20]uint16)(unsafe.Pointer(p))[:n:n]
	return syscall.UTF16ToString(slice)
}
