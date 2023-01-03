package index

import (
	"log"

	"github.com/OneOfOne/xxhash"
	"github.com/sp4ke/hashmap"
)

type Index = *hashmap.RBTree
type HashTree = *hashmap.RBTree

// In memory index used for fast lookup of url->node pairs
// to quickly detect bookmark which changed when bookmarks are reloaded
// from browser on a watch event
// Input `in` must be of type []byte
// The index is a map of [urlhash]*Node
func xxHashFunc(in interface{}) uint64 {
	input, ok := in.(string)
	if !ok {
		log.Panicf("wrong data type to hash, exptected string given %T", in)
	}
	sum := xxhash.ChecksumString64(input)
	//log.Debugf("Calculating hash of %s as %d", input, sum)
	return sum
}

// Returns  *hashmap.RBTree
func NewIndex() *hashmap.RBTree {
	return hashmap.New(xxHashFunc)
}
