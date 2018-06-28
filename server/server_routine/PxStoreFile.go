package server_routine

import (
	"strconv"
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


type PxStoreFile struct {
	pvcpath   map[string]string
	hashtable *consistenthash.Map
	writerMap *MyMap
}

func NewPxStoreFile() *PxStoreFile {
	px := &PxStoreFile{
		pvcpath: make(map[string]string),
		hashtable: consistenthash.New(1, nil),
		writerMap: NewMap(),
	}

	for i := 1; i < 10; i++ {
		pvc := "volume" + strconv.Itoa(i)
		path := "/tmp/test-portworx-volume" + strconv.Itoa(i)
		px.pvcpath[pvc] = path
	}

	for _,v := range px.pvcpath {
		px.hashtable.Add(v)
	}

	return px
}


//package drop , taking too long
func (px *PxStoreFile) StoreFile(ctx context.Context, fileName string, reader io.Reader) (int64, error) {
	writer := px.writerMap.Value(fileName)
	fmt.Println(fileName)

	if writer == nil {
		//create file
		volpath := px.hashtable.Get(fileName)
		fmt.Printf("the volume path get from consistent hash table: %s\n", volpath)
		if _, err := os.Stat(volpath); os.IsNotExist(err) {
			err := os.MkdirAll(volpath, os.ModePerm)
			if err != nil {
				return -1, errors.New("Create directory fails")
			}
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

	n, err := io.Copy(f, reader)// whether this stream bufio go by example writing files
	if err != nil {
		px.writerMap.Value(fileName).Close()
		return -3, errors.New("fail when writing to file")
	}
	return n, nil
}

func (px *PxStoreFile) CloseFile(fileName string) error {
	px.writerMap.Value(fileName).Close()
	return nil
}