package server_routine

import (

	"github.com/golang/groupcache/consistenthash"
	"io"
	"sync"
	"fmt"
	"os"
	"errors"
	"path/filepath"
	"context"
)

type MyMap struct {
	mu sync.Mutex
	m map[string]io.WriteCloser
}

func NewMap() *MyMap {
	mymap := &MyMap{
		m: make(map[string]io.WriteCloser),
	}
	return mymap
}

func (ma *MyMap) Add(k string , v io.WriteCloser) {
	ma.mu.Lock()
	ma.m[k] = v
	ma.mu.Unlock()
}

func (ma *MyMap) Value(k string) (v io.WriteCloser) {
	ma.mu.Lock()
	_, exist := ma.m[k]
	if exist == true {
		v = ma.m[k]
	} else {
		v = nil
	}
	ma.mu.Unlock()
	return v
}


type PxFileManager struct {
	pvcpath   map[string]string
	hashtable *consistenthash.Map
	writerMap *MyMap
}

func NewPxFileManager(deploymentName string) *PxFileManager {
	px := &PxFileManager{
		pvcpath: make(map[string]string),
		hashtable: consistenthash.New(1, nil),
		writerMap: NewMap(),
	}
	pvcs, err := GetPvcPath(deploymentName)
	if err != nil {
		fmt.Println(err)
	}

	for k, v:= range pvcs {
		fmt.Printf("the pvc is %v, the pvc paht is %v\n", k, v)
		px.pvcpath[k] = v

	}

	for _,v := range px.pvcpath {
		px.hashtable.Add(v)
	}
	fmt.Println("============new function ==============")
	for k, v := range px.pvcpath {
		fmt.Printf("the pvc is %v, and the path is %v\n", k, v)
	}

	return px
}


//package drop , taking too long
// This function find the corresponding consistent hash directory and
// write the content in the reader to the file in the corresponding directory.
// If the file or directory doesn't exit, it create one and add to writeMap.
//
func (px *PxFileManager) StoreFile(ctx context.Context, fileName string, reader io.Reader) (int64, error) {
	writer := px.writerMap.Value(fileName)
	fmt.Println(fileName)

	if writer == nil {
		//create file
		volpath := px.hashtable.Get(fileName)
		fmt.Printf("the volume path get from consistent hash table: %s\n", volpath)
		if _, err := os.Stat(volpath); os.IsNotExist(err) {
			return -4, errors.New("The path get from consistent hash table not exist")
			//err := os.MkdirAll(volpath, os.ModePerm)
			//if err != nil {
			//	return -1, errors.New("Create directory fails")
			//}
		}
		newPath := filepath.Join(volpath, fileName)
		f, err := os.Create(newPath)

		if err != nil {
			f.Close()
			return -2, errors.New("Create file fails")
		}
		//add to map
		px.writerMap.Add(fileName, f)
	}
	f := px.writerMap.Value(fileName)

	// write to file.

	n, err := io.Copy(f, reader) // whether this stream bufio go by example writing files
	if err != nil {
		px.writerMap.Value(fileName).Close()
		return -3, errors.New("fail when writing to file")
	}
	return n, nil
}