package server_routine

import (
	"net"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pb "github.com/LynneD/TransferFile/transferfile"
	"os"
	"io"
	"crypto/md5"
	"fmt"
	"errors"
	"sync"
	"path/filepath"
)

const clientfiledir = "tmp/clientfile"

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
			log.WithFields(log.Fields{fileName: sendFileRequest.Md5Sum}).Info("stream receiving file data fails")
			return err
		}

		chechSum := md5.Sum(sendFileRequest.Data)
		md5str := fmt.Sprintf("%x", chechSum)
		if sendFileRequest.Md5Sum != md5str {
			writerMap.Value(fileName).Close()
			log.WithFields(log.Fields{fileName:md5str}).Info("Wrong file data received")
			return errors.New("wrong file data.")
		}

		fmt.Println(md5str)
		md5array = append(md5array, md5str)

		fileName = filepath.Base(sendFileRequest.FileName)
		f := openFile(fileName)

		if f == nil {
			return errors.New("creating file fails")
		}
		// write to file.
		n, err := f.Write(sendFileRequest.Data)
		if err != nil {
			writerMap.Value(fileName).Close()
			log.WithFields(log.Fields{fileName:"Write"}).Info("Writing to file fails")
			return errors.New("fail when writing to file")
		}
		sum += int64(n)

	}
}

func openFile(fileName string) (io.Writer){
	writer := writerMap.Value(fileName)
	fmt.Println(fileName)

	if writer == nil {
		//create file

		newPath := filepath.Join(clientfiledir, fileName)
		f, err := os.Create(newPath)
		//f, err := os.Create(fileName)
		if err != nil {
			f.Close()
			log.WithFields(log.Fields{fileName:"create file"}).Info("Creating file fails")
			return nil
		}
		//add to map
		writerMap.Add(fileName, f)
	}
	return writerMap.Value(fileName)
}


//filepath.Split
//filepath.Join
func ServerRoutine(host string, port string) {
	if _, err := os.Stat(clientfiledir); os.IsNotExist(err) {
		err := os.Mkdir(clientfiledir, os.ModePerm)
		if err != nil {
			log.WithFields(log.Fields{clientfiledir:"Create Dir"}).Info("Creating directory fails")
		}
	}

	writerMap.m = make(map[string]io.WriteCloser)
	lis, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalf("fail to listen on port%v, %v", port, err)
	}
	s := grpc.NewServer()
	pb.RegisterTransferFileServer(s, &myServer{})
	s.Serve(lis)
}
