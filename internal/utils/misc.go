//
// Copyright (c) 2023-2025 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify it under the terms of
// the GNU Affero General Public License as published by the Free Software Foundation,
// either version 3 of the License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
// PURPOSE.  See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License along with
// gosuki.  If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"math/rand"
	"strings"
)

func CamelCase(in string) string {
	words := strings.Fields(in)
	if len(words) == 0 {
		return ""
	}
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if len(words[i]) == 0 {
			continue
		}
		first := strings.ToUpper(string(words[i][0]))
		rest := words[i][1:]
		result += first + rest
	}
	return result
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
func UseVar(any any) {}

// function that iterates through the list of string, for each element it
// replaces the occurence of old with new, and returns the updated list
func ReplaceInList(src []string, old string, new string) []string {
	lenSrc := len(src)
	result := make([]string, lenSrc)
	for i, s := range src {
		result[i] = strings.ReplaceAll(s, old, new)
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
func Map[T, U comparable](f func(item T) U, list []T) []U {
	var newList []U
	for _, v := range list {
		newList = append(newList, f(v))
	}
	return newList
}
