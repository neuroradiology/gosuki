package utils


// Return string from slice of bytes
func S(value interface{}) string {
	return string(value.([]byte))
}

// Extends a slice of T with element `in`, like a Set
func Extends[T comparable](list []T, in T) []T {
	for _, val := range list {
		if in == val {
			return list
		}
	}
	return append(list, in)
}

// Return true if elm in list
func Inlist[T comparable](list []T, elm T) bool {
	for _, v := range list {
		if elm == v {
			return true
		}
	}

	return false
}

