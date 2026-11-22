//go:build windows && amd64

package edge

import "testing"

func TestCallbackRootsRetainAndRelease(t *testing.T) {
	const ptr = uintptr(1)
	handler := new(int)

	retainCallbackRoot(ptr, handler)

	callbackRoots.Lock()
	got, ok := callbackRoots.handlers[ptr]
	callbackRoots.Unlock()
	if !ok || got != handler {
		t.Fatal("callback root was not retained")
	}

	releaseCallbackRoot(ptr)

	callbackRoots.Lock()
	_, ok = callbackRoots.handlers[ptr]
	callbackRoots.Unlock()
	if ok {
		t.Fatal("callback root was not released")
	}
}

func TestPendingScriptCallbackRootSurvivesDestroy(t *testing.T) {
	c := NewChromium(0)
	handler := NewExecuteScriptCompletedHandler(func(string) {})
	ptr := handler.AsPtr()

	c.scriptHandlers[1] = handler
	callbackAddRef(&handler.refs)
	retainCallbackRoot(ptr, handler)
	callbackReleaseAndUnroot(ptr, &handler.refs)

	c.Destroy()

	callbackRoots.Lock()
	got, ok := callbackRoots.handlers[ptr]
	callbackRoots.Unlock()
	if !ok || got != handler {
		t.Fatal("Destroy released a pending script callback root")
	}

	callbackReleaseAndUnroot(ptr, &handler.refs)

	callbackRoots.Lock()
	_, ok = callbackRoots.handlers[ptr]
	callbackRoots.Unlock()
	if ok {
		t.Fatal("callback root remained after the COM reference count reached zero")
	}
}
