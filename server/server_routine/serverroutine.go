package server_routine

import (
	"net"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pb "github.com/LynneD/TransferFile/transferfile"
	"io"
	"crypto/md5"
	"fmt"
	"errors"
	"path/filepath"
	"bytes"
	"context"
	"time"
)

var FileManage FileManager

type myServer struct {
}

func (s *myServer) StoreSourceFile(stream pb.TransferFile_StoreSourceFileServer) error {
	var sum int64 = 0
	var md5array []string
	var fileName string
	ctx := stream.Context()

	deadline, ok := ctx.Deadline()
	if !ok {
		log.WithFields(log.Fields{"server": "refuse request"}).Info(
			"Client send a context with no deadline")
		return errors.New("Client send a context with no deadline.")
	}

	s_ctx, cancel1 := context.WithTimeout(context.Background(), deadline.Sub(time.Now()))
	defer cancel1()

	for {
		if s_ctx.Err() == context.Canceled {
			log.WithFields(log.Fields{"server routine": "Timeout"}).Info(
				"The context is done. Server routine timeout")
			return errors.New("The context is timed out.")
		}
		sendFileRequest, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("io.EOF")
			//writerMap.Value(fileName).Close()
			sendFileResponse := &pb.SendFileResponse{BytesWritten: sum, Md5Sum: md5array}
			return stream.SendAndClose(sendFileResponse)
		}
		if err != nil {
			//writerMap.Value(fileName).Close()
			log.WithFields(log.Fields{"Stream":"Error happens when receiving data from grpc stream"}).Info(err)
			return err
		}

		chechSum := md5.Sum(sendFileRequest.Data)
		md5str := fmt.Sprintf("%x", chechSum)
		if sendFileRequest.Md5Sum != md5str {
			//writerMap.Value(fileName).Close()
			log.WithFields(log.Fields{fileName: md5str}).Info("Wrong file data received")
			return errors.New("wrong file data.")
		}

		fmt.Println(md5str)
		md5array = append(md5array, md5str)


		fileName = filepath.Base(sendFileRequest.FileName)
		//fmt.Printf("the filepath base is : %s \n", fileName)

		r := bytes.NewReader(sendFileRequest.Data)
		//fmt.Printf("the io.reader is %v\n", r)
		n, err := FileManage.StoreFile(s_ctx, fileName, r)

		if err != nil {
			if n == -1 {
				log.WithFields(log.Fields{"pvcpath": "Create Dir"}).Info("Creating directory fails")
			} else if n == -2 {
				log.WithFields(log.Fields{fileName: "create file"}).Info("Creating file fails")
			} else if n == -3 {
				log.WithFields(log.Fields{fileName: "Write"}).Info("Writing to file fails")
			} else if n == -4 {
				log.WithFields(log.Fields{fileName: "doesn't have this path"}).Info("The path get from consistent hash table not exist")
			}
			return errors.New("storing file fails")
		}
		sum += n
	}

}

func (s *myServer) EnumerateFiles(ctx context.Context, request *pb.GetFileListRequest) (*pb.GetFileListResponse, error) {
	var filelist []string
    return &pb.GetFileListResponse{FileName:filelist}, nil
}

func (s *myServer) GetPVC(ctx context.Context, request *pb.GetPVCRequest) (*pb.GetPVCResponse, error) {
    return &pb.GetPVCResponse{}, nil
}

func (s *myServer) DistributeResults(stream pb.TransferFile_DistributeResultsServer) error {
	return nil
}

func ServerRoutine(host string, port string, volumeProvider string, deploymentName string) {

	if volumeProvider == "portworx" {
		FileManage = NewPxFileManager(deploymentName)
	} else if volumeProvider == "test" {
		return
	}

	if FileManage == nil {
		log.WithFields(log.Fields{"Init":"Fail create PX"}).Info(
			"Fail to create portworx storage file instance")
		return
	}

	lis, err := net.Listen("tcp", host +":" + port)
	if err != nil {
		log.Fatalf("fail to listen on %v:%v, %v", port, err)
	}
	s := grpc.NewServer()
	pb.RegisterTransferFileServer(s, &myServer{})
	s.Serve(lis)
}

