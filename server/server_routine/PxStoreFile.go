package server_routine

import "strconv"

type PxStoreFile struct{
	pvcpath map[string]string
}

func NewPxStoreFile() *PxStoreFile {
	px := new(PxStoreFile)
	for i := 1; i < 10; i++ {
		pvc := "volume" + strconv.Itoa(i)
		path := "/tmp/test-portworx-volume" + strconv.Itoa(i)
		px.pvcpath[pvc] = path
	}
	for _, v := range px.pvcpath {
		hashtable.Add(v)
	}
	return px
}

