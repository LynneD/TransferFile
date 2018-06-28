package server_routine

import (
	"fmt"
	"bytes"
	"testing"
)

// golang 4test
func TestNewPxStoreFile (t *testing.T) {
	px := NewPxStoreFile()
	for k, v := range px.pvcpath {
		fmt.Printf("the pvc is %v, the corresponding path is %v\n", k, v)
		fmt.Printf("the path get from hash table is %v\n", px.hashtable.Get(v))
	}
}

func TestStoreFile (t *testing.T) {
	px := NewPxStoreFile()
	data := []byte{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', '!'}
	r := bytes.NewReader(data)
	px.StoreFile("testcode.txt", r)
}