package utils

// Return string from slice of bytes
func S(value interface{}) string {
	return string(value.([]byte))
}

func Extends(list []string, in string) []string {
	for _, val := range list {
		if in == val {
			return list
		}
	}
	return append(list, in)
}

// Return true if elm in list
func Inlist(list []string, elm string) bool {
	for _, v := range list {
		if elm == v {
			return true
		}
	}

	return false
}
