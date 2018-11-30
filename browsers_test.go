package main

import "testing"

func TestShutdown(t *testing.T) {
	t.Error("abondon all db operations on Shutdown()")
}

func TestOnShutdownCloseWatcherThread(t *testing.T) {
	t.Error("Close Watcher Thread on Shutdown")
}

func TestBrowserInterface(t *testing.T) {
	t.Error("Check Browser required interface")
}
