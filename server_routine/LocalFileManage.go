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
	"strconv"
	"regexp"
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


type LocalFileManager struct {
	path   map[string]string
	hashtable *consistenthash.Map
	writerMap *MyMap
	distributedFileList map[string]string // fileName : pvc
}

func NewLocalFileManager() *LocalFileManager {
	lc := &LocalFileManager{
		path: make(map[string]string),
		hashtable: consistenthash.New(1, nil),
		writerMap: NewMap(),
		distributedFileList: make(map[string]string),
	}


	for i := 1; i < 6; i++ {
		vol := "volume" + strconv.Itoa(i)
		dir := "/tmp/test-px-" + vol
		lc.path[vol] = dir
	}

	for k,_ := range lc.path {
		lc.hashtable.Add(k)
	}
	fmt.Println("============new function ==============")
	for k, v := range lc.path {
		fmt.Printf("the pvc is %v, and the path is %v\n", k, v)
	}

	return lc
}


//package drop , taking too long
// This function find the corresponding consistent hash directory and
// write the content in the reader to the file in the corresponding directory.
// If the file or directory doesn't exit, it create one and add to writeMap.
//
func (lc *LocalFileManager) StoreFile(ctx context.Context, fileName string, reader io.Reader) (int64, error) {
	if _, ok := lc.distributedFileList[fileName]; ok {
		return 0, nil
	}
	writer := lc.writerMap.Value(fileName)
	//fmt.Println(fileName)
	// if already finish return finish cannot write already have this file
	if writer == nil {
		//create file
		vol := lc.hashtable.Get(fileName)
		volpath := lc.path[vol]
		fmt.Printf("the volume path get from consistent hash table: %s\n", volpath)
		if _, err := os.Stat(volpath); os.IsNotExist(err) {
			//return -4, errors.New("The path get from consistent hash table not exist")
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
		lc.writerMap.Add(fileName, f)
	}
	f := lc.writerMap.Value(fileName)

	// write to file.

	n, err := io.Copy(f, reader) // whether this stream bufio go by example writing files
	if err != nil {
		lc.writerMap.Value(fileName).Close()
		return -3, errors.New("fail when writing to file")
	}
	return n, nil
}

func (lc *LocalFileManager) GetFileList(ctx context.Context, regexp *regexp.Regexp) ([]string, error) {
	fileList := make([]string, 0)
	for k, _ := range lc.distributedFileList {
		match := regexp.MatchString(k)
		if match {
			fileList = append(fileList, k)
		}
	}
	return fileList, nil
}

/**

 */
func (lc *LocalFileManager) GetPVCList(ctx context.Context, fileList []string) ([]string, error) {
	pvcList := make([]string, 0)
	for _, file := range fileList {
		if pvc, ok := lc.distributedFileList[file]; ok {
			pvcList = append(pvcList, pvc)
		}
	}
	return pvcList, nil
}

/**

 */
func (lc *LocalFileManager) Clean(isDone bool, fileName string) {
	if isDone == true {
		// add to distributedFileList
		vol := lc.hashtable.Get(fileName)
		lc.distributedFileList[fileName] = vol
	}
	lc.writerMap.Value(fileName).Close()
}