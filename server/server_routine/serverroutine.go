package server_routine

import (
	"net"
	"log"
	"google.golang.org/grpc"
	pb "TransferFile/transferfile"
	"os"
	"io"
	"crypto/md5"
	"fmt"
	"errors"
	"sync"
	"time"
)



type myMap struct {
	mu sync.Mutex
	m map[string]io.WriteCloser
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

var writerMap myMap

type myServer struct {
}

func (s *myServer) SendFile(stream pb.TransferFile_SendFileServer) error {
	var sum int64 = 0
	var md5array []string
	var fileName string

	for {
		//receive data
		sendFileRequest, err := stream.Recv()

		if err == io.EOF {
			fmt.Println("io.EOF")
			writerMap.Value(fileName).Close()
			sendFileResponse := &pb.SendFileResponse{BytesWritten:sum, Md5Sum:md5array}
			return stream.SendAndClose(sendFileResponse)
		}
		if err != nil {
			writerMap.Value(fileName).Close()
			log.Fatal("stream recv data fail")
			return err
		}

		chechSum := md5.Sum(sendFileRequest.Data)
		md5str := fmt.Sprintf("%x", chechSum)
		if sendFileRequest.Md5Sum != md5str {
			writerMap.Value(fileName).Close()
			return errors.New("wrong file data.")
		}
		fmt.Println(md5str)
		md5array = append(md5array, md5str)

		fileName = sendFileRequest.FileName
		f := openFile(fileName)

		// write to file.
		n, err := f.Write(sendFileRequest.Data)
		if err != nil {
			writerMap.Value(fileName).Close()
			return err
		}
		sum += int64(n)
		time.Sleep(5*time.Second)
	}
}

func openFile(fileName string) (io.Writer){
	writer := writerMap.Value(fileName)
	fmt.Println(fileName)

	if writer == nil {
		//create file
		f, err := os.Create(fileName)
		if err != nil {
			f.Close()
			log.Fatalf("create file %v failed: %v", fileName, err)
		}
		//add to map
		writerMap.Add(fileName, f)
	}
	return writerMap.Value(fileName)
}



func ServerRoutine(host string, port string) {
	writerMap.m = make(map[string]io.WriteCloser)
	lis, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalf("fail to listen on port%v, %v", port, err)
	}
	s := grpc.NewServer()
	pb.RegisterTransferFileServer(s, &myServer{})
	s.Serve(lis)

}
