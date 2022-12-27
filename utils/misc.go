package utils

import (
	"math/rand"
	"strings"
)

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


// Use to shutoff golang "unused variable comment"
func UseVar(any interface{}) {
    return
}

// function that iterates through the list of string, for each element it 
// replaces the occurence of old with new, and returns the updated list 
func ReplaceInList(l []string, old string, new string) []string {
    var result []string
    for _, s := range l {
        result = append(result, strings.Replace(s, old, new, -1))
    }
    return result
}

// Generate a unique random string with the specified length
func GenStringID(n int) string {
    var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    b := make([]rune, n)
    for i := range b {
        b[i] = letter[rand.Intn(len(letter))]
    }
    return string(b)
}
