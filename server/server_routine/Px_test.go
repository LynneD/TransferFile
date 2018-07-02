package server_routine

import (
	"fmt"
	"bytes"
	"testing"
	"context"
)

//
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
	ctx := context.Background()


	testCases := []string{
		"test1.txt",
		"test2.txt",
		"test3.txt",
		"test4.txt",
	}

	for _, s := range testCases {
		px.StoreFile(ctx, s, r)



	}
}