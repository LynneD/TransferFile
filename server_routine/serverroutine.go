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
	"regexp"
)


var FileManage FileManager


type myServer struct {
}

/**
This is a client-side stream grpc function.
StoreSourceFile get the file chunk(filename, file chunk data, hash value) from client and
store the file into one of the storage volumes
*/
func (s *myServer) StoreSourceFile(stream pb.TransferFile_StoreSourceFileServer) error {
	fmt.Println("=====================store file=====================")
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
		sendFileRequest, err := stream.Recv()

		if s_ctx.Err() == context.Canceled {
			fmt.Println("context canceled")
			FileManage.Clean(false, fileName)

			log.WithFields(log.Fields{"server routine": "Timeout"}).Info(
				"The context is done. Server routine timeout")
			return errors.New("The context is timed out.")
		}
		if err == io.EOF {
			fmt.Println("io.EOF")

			FileManage.Clean(true, fileName)

			sendFileResponse := &pb.SendFileResponse{BytesWritten: sum, Md5Sum: md5array}

			return stream.SendAndClose(sendFileResponse)
		}
		if err != nil {
			fmt.Println("Other stream error")
			FileManage.Clean(false, fileName)

			log.WithFields(log.Fields{"Stream":"Error happens when receiving data from grpc stream"}).Info(err)
			return err
		}

		fileName = filepath.Base(sendFileRequest.FileName)
		fmt.Println("The file Name get from client is " + fileName)
		// use hash value to check the correctness of file data
		md5sum, err := checkFileData(sendFileRequest.Data, sendFileRequest.Md5Sum)
		if err != nil {
			fmt.Println("Check data: data is wrong")
			FileManage.Clean(false, fileName)

			log.WithFields(log.Fields{fileName: "checkdata error"}).Info(err)
			return err
		}
		md5array = append(md5array, md5sum)

		// store file data
		fmt.Println("Store data")
		r := bytes.NewReader(sendFileRequest.Data)
		n, err := FileManage.StoreFile(s_ctx, fileName, r)
		//if n == 0{
		//	continue
		//}
		if err != nil {
			fmt.Println("PXFileManager store file fail")
			FileManage.Clean(false, fileName)

			log.WithFields(log.Fields{fileName: "FileManage store file"}).Info(err)
			return errors.New("storing file fails")
		}

		sum += n
	}

}

/**
EnumerateFiles get the regular expression from client and returns all matched files distributed on the
storage volumes
 */
func (s *myServer) EnumerateFiles(ctx context.Context, request *pb.GetFileListRequest) (*pb.GetFileListResponse, error) {
	fmt.Println("==========================Enum Files==========================")
	reg := request.RegExp
	fmt.Printf("the regexp is %v\n", reg)
	r, err := regexp.Compile(reg)
	if err != nil {
		log.WithFields(log.Fields{reg: "FileManage enumerateFiles"}).Info(
			"The regular expression is not right")
	}
	fileList, err := FileManage.GetFileList(ctx, r) // ? can assign like this
	fmt.Printf("the file List is %v\n", fileList)
	if err != nil {
		log.WithFields(log.Fields{reg: "FileManage enumerateFiles"}).Info(
			"Get File List error")
	}
    return &pb.GetFileListResponse{FileName:fileList}, nil
}

/**
GetPVC get a list of files from client and returns the corresponding list of pvc on which the files store
 */
func (s *myServer) GetPVC(ctx context.Context, request *pb.GetPVCRequest) (*pb.GetPVCResponse, error) {
	fmt.Println("======================Get PVC===========================")
	fmt.Printf("the file name is %v\n",request.FileName)
	pvcList, err := FileManage.GetPVCList(ctx, request.FileName) // ? can assign like this
	if err != nil {
		log.WithFields(log.Fields{"GetPVC": "FileManage GetPVC"}).Info(
			"Get PVC List error")
	}
	fmt.Printf("the pvc list is %v\n", pvcList)
    return &pb.GetPVCResponse{Pvc:pvcList}, nil
}


/**

 */
func (s *myServer) DistributeResults(stream pb.TransferFile_DistributeResultsServer) error {
	return nil
}


/**
Check whether the fileData's md5 hash value matchs the md5sum string.
The function returns a bool value indicates whether the data is correct
 */
func checkFileData(fileData []byte, md5sum string) (string, error) {
	chechSum := md5.Sum(fileData)
	md5str := fmt.Sprintf("%x", chechSum)
	if md5sum != md5str {
		return "", errors.New("Wrong file data received")
	}
	return md5str, nil
}


func ServerRoutine(host string, port string, volumeProvider string, deploymentName string) {
	if volumeProvider == "portworx" {
		fmt.Println("portworx")
		FileManage = NewPxFileManager(deploymentName)
	} else if volumeProvider == "local" {
		FileManage = NewLocalFileManager()
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

