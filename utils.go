package main

// Return string from slice of bytes
func s(value interface{}) string {
	return string(value.([]byte))
}
