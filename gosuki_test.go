package main_test

import "testing"

// effectively tests db/fn: initDB()
// TEST: test that gosuki.db is properly created on startup
// implement this as integration test
func TestGosukiDBCreatedAtStartup(t *testing.T) {
	t.Error("[integration] if gosuki.db does not exist create it")
}
