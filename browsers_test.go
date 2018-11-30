package main

import "testing"

func TestOnShutdown(t *testing.T) {
	t.Error("abondon all db operations on Shutdown()")
}

func TestOnShutdownCloseWatcherThread(t *testing.T) {
	t.Error("Close Watcher Thread on Shutdown")
}
