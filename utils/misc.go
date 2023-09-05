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
func Extends[T comparable](list []T, in ...T) []T {
	for _, val := range in {
        if !InList(list, val) {
            list = append(list, val)
        }
	}
	return list
}

// Return true if elm in list
func InList[T comparable](list []T, elm T) bool {
	for _, v := range list {
		if elm == v {
			return true
		}
	}

	return false
}


// Use to shutoff golang "unused variable comment"
func UseVar(any interface{}) {}

// function that iterates through the list of string, for each element it 
// replaces the occurence of old with new, and returns the updated list 
func ReplaceInList(l []string, old string, new string) []string {
    var result []string
    for _, s := range l {
        result = append(result, strings.ReplaceAll(s, old, new, ))
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


// map takes a list and a function and returns a new list
func Map[T , U comparable](f func(item T) U, list []T) []U {
    var newList []U
    for _, v := range list {
        newList = append(newList, f(v))
    }
    return newList
}
