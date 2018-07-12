package server_routine

import (
	"fmt"
	"testing"
	"bytes"
)

//

func TestLocalFileManage(t *testing.T) {
	lc := NewLocalFileManager()
	for k, v := range lc.pvcpath {
		fmt.Printf("the pvc is %v, the corresponding path is %v\n", k, v)
		fmt.Printf("the path get from hash table is %v\n", lc.hashtable.Get(v))
	}

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
