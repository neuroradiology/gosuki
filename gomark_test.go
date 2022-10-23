package main_test

import "testing"

// effectively tests db/fn: initDB()
// TEST: test that gomark.db is properly created on startup
// implement this as integration test
func TestGomarkDBCreatedAtStartup(t *testing.T) {
	t.Error("[integration] if gomark.db does not exist create it")
}
