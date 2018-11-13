package main

// Return string from slice of bytes
func s(value interface{}) string {
	return string(value.([]byte))
}

func extends(list []string, in string) []string {
	for _, val := range list {
		if in == val {
			return list
		}
	}
	return append(list, in)
}
