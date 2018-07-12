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
	"regexp"
)

type myMap struct {
	mu sync.Mutex
	m map[string]io.WriteCloser
}

func newMap() *myMap {
	mymap := &myMap{
		m: make(map[string]io.WriteCloser),
	}
	return mymap
}

func (ma *myMap) Add(k string , v io.WriteCloser) {
	ma.mu.Lock()
	ma.m[k] = v
	ma.mu.Unlock()
}

func (ma *myMap) Value(k string) (v io.WriteCloser) {
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
	pvcpath   map[string]string           // pvc : mountpath
	hashtable *consistenthash.Map         // pvc
	writerMap *myMap                      // fileName : io.Writer
	distributedFileList map[string]string // fileName : pvc
}

func NewPxFileManager(deploymentName string) *PxFileManager {
	px := &PxFileManager{
		pvcpath: make(map[string]string),
		hashtable: consistenthash.New(1, nil),
		writerMap: newMap(),
		distributedFileList: make(map[string]string),
	}
	pvcs, err := GetPvcPath(deploymentName)
	if err != nil {
		fmt.Println(err)
	}

	for k, v:= range pvcs {
		//fmt.Printf("the pvc is %v, the pvc paht is %v\n", k, v)
		px.pvcpath[k] = v

	}

	for k,_ := range px.pvcpath {
		px.hashtable.Add(k)
	}
	fmt.Println("============The portworx volume info==============")
	for k, v := range px.pvcpath {
		fmt.Printf("the pvc is %v, and the path is %v\n", k, v)
	}

	return px
}



// This function find the corresponding consistent hash directory and
// write the content in the reader to the file in the corresponding directory.
// If the file or directory doesn't exit, it create one and add to writeMap.
//
func (px *PxFileManager) StoreFile(ctx context.Context, fileName string, reader io.Reader) (int64, error) {
	//if _, ok := px.distributedFileList[fileName]; ok {
	//	return 0, nil
	//}
	fmt.Println("=============PXFileManager StoreFile===============")
	writer := px.writerMap.Value(fileName)
	//fmt.Println(fileName)

	if writer == nil {
		fmt.Println("doesn't have this file " + fileName )
		//create file
		vol := px.hashtable.Get(fileName)
		volpath := px.pvcpath[vol]
		fmt.Printf("the volume path get from consistent hash table: %s\n", volpath)
		if _, err := os.Stat(volpath); os.IsNotExist(err) {
			return -1, errors.New("The path get from consistent hash table not exist")
		}
		newPath := filepath.Join(volpath, fileName)
		f, err := os.Create(newPath)

		if err != nil {
			fmt.Println("Create file fails")
			f.Close()
			return -1, errors.New("Create file fails")
		}
		//add to map
		px.writerMap.Add(fileName, f)
	}
	f := px.writerMap.Value(fileName)

	// write to file.
	fmt.Println("write to file")
	n, err := io.Copy(f, reader) // whether this stream bufio go by example writing files
	if err != nil {
		fmt.Println("wirte to file fail")
		px.writerMap.Value(fileName).Close()
		return -1, errors.New("fail when writing to file")
	}
	return n, nil
}

/**

 */
func (px *PxFileManager) GetFileList(ctx context.Context, regexp *regexp.Regexp) ([]string, error) {
	fmt.Println("===================PXFileManager get file list=====================")
	fileList := make([]string, 0)
	for k, _ := range px.distributedFileList {
		//fmt.Printf("the file name is %v, the volume is %v\n", k, v)
		match := regexp.MatchString(k)
		if match {
			fileList = append(fileList, k)
		}
	}
	return fileList, nil
}

/**

 */
func (px *PxFileManager) GetPVCList(ctx context.Context, fileList []string) ([]string, error) {
	fmt.Println("===================PXFileManager get pvc list=====================")
	pvcList := make([]string, 0)
	for _, file := range fileList {
		if pvc, ok := px.distributedFileList[file]; ok {
			pvcList = append(pvcList, pvc)
		}
	}
	return pvcList, nil
}

/**

 */
func (px *PxFileManager) Clean(isDone bool, fileName string) {
	fmt.Println("===============PXFileMnager clean============")
	if isDone == true {
		// add to distributedFileList
		vol := px.hashtable.Get(fileName)
		px.distributedFileList[fileName] = vol
		fmt.Printf("the file store is %v, the distributed volume is %v\n", fileName, vol)
	} //else delete the file
	px.writerMap.Value(fileName).Close()
}