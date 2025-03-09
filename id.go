package main

import (
	"crypto/md5"
	"fmt"
	"math/rand"
)

// newID generates a new ID based on the input.
// If none given, generate random ID.
func newID(in ...interface{}) []byte {
	if in == nil || len(in) == 0 {
		in = []interface{}{rand.Int63()}
	}

	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprint(in...)))

	sum := hasher.Sum(nil)
	return sum
}
