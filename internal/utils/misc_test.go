package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInlist(t *testing.T) {

	t1 := []int{1, 2, 3, 4}
	assert.Equal(t, true, InList(t1, 4))
	assert.Equal(t, false, InList(t1, 5))

	t2 := []string{"one", "two", "three"}
	assert.Equal(t, true, InList(t2, "three"))
	assert.Equal(t, false, InList(t2, "five"))
}

func TestExtends(t *testing.T) {
	t1 := []int{1, 2, 3}
	assert.Equal(t, []int{1, 2, 3, 4}, Extends(t1, 4))
	assert.NotEqual(t, []int{1, 2, 3, 3}, Extends(t1, 3))
}
