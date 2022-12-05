package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInlist(t *testing.T) {

	t1 := []int{1, 2, 3, 4}
	assert.Equal(t, true, Inlist(t1, 4))
	assert.Equal(t, false, Inlist(t1, 5))

	t2 := []string{"one", "two", "three"}
	assert.Equal(t, true, Inlist(t2, "three"))
	assert.Equal(t, false, Inlist(t2, "five"))
}

func TestExtends(t *testing.T) {
	t1 := []int{1, 2, 3}
	assert.Equal(t, []int{1, 2, 3, 4}, Extends(t1, 4))
	assert.NotEqual(t, []int{1, 2, 3, 3}, Extends(t1, 3))
}
