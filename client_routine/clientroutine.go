package clientroutine

import (
	"google.golang.org/grpc"
	"os"
	"log"
	pb "github.com/LynneD/TransferFile/transferfile"
	"context"
	"time"
	"io"
	"fmt"
	"crypto/md5"
	"reflect"
	"path/filepath"
)


var filename []string

var length int
var index int

func sendFile(client pb.TransferFileClient) {
	var md5array []string
	ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)// set as a parameter
	defer cancel()

	stream, err := client.StoreSourceFile(ctx)
	if err != nil {
		log.Fatalf("%v.SendFile(_) = _, %v", client, err)
	}


	f, err := os.Open(filename[index])
	if err != nil {
		log.Fatalf("Open file %v failed, %v", filename, err)
	}
	defer f.Close()

	// store data waiting to be sent
	buf := make([]byte, length)
	var errorCount int64 = 0

	// read chunk from file and send to server
	for {
		if errorCount > 1 {
			log.Fatalf("Failed sending file")
			return
		}

		n, err := f.Read(buf)
		if err == io.EOF {// Read to the end of the file
			break
		}
		if err != nil {
			log.Fatalf("read from file %v failed, %v", filename, err)
		}

		chechSum := md5.Sum(buf[:n])
		md5str := fmt.Sprintf("%x", chechSum)
		md5array = append(md5array, md5str)

		//send
		sendFileRequest := &pb.SendFileRequest{FileName: filename[index], Md5Sum:md5str, Data:buf[:n]}
		if err := stream.Send(sendFileRequest); err != nil {
			if err == io.EOF {
				break
			}
			errorCount++
			log.Fatalf("%v.Send(%v), error: %v", stream, sendFileRequest.FileName, err)
		} else {
			errorCount = 0
		}
		//time.Sleep(5*time.Second)
	}

	// close stream and receive result from server
	sendFileResponse, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v", stream, err, nil)
	}
	if reflect.DeepEqual(md5array, sendFileResponse.Md5Sum) {
		log.Printf("Store %s completed                   %d bytes written\n",
			filepath.Base(filename[index]), sendFileResponse.BytesWritten)
	} else {
		log.Fatalf("Wrong data")
	}

}

func enumerateFiles(client pb.TransferFileClient, regexp string) []string {
	//fmt.Println("===================enumerateFiles=====================")
	//fmt.Println("The regular expression input is: " + regexp)
	ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancel()

	response, err := client.EnumerateFiles(ctx, &pb.GetFileListRequest{RegExp:regexp})
	if err != nil {
		log.Fatal("%v.EnumerateFiles(_) = _, %v: ", client, err)
	}
	fmt.Printf("the file list is %v\n", response)
	return response.FileName
}

func getPVC(client pb.TransferFileClient, fileList []string) {
	//fmt.Println("===================getPVC=====================")
	//fmt.Printf("The file list is %v\n", fileList)
	ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancel()

	response, err := client.GetPVC(ctx, &pb.GetPVCRequest{FileName:fileList})
	if err != nil {
		log.Fatal("%v.GetPVC(_) = _, %v: ", client, err)
	}
	fmt.Printf("the pvc list is %v\n", response)
}

func StoreFile(chunksize int, host string, port string, fileName []string) {
	address := host + ":" + port
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to server %v, %v", address, err)
	}

	length = chunksize

	filename = fileName   // ? can assign value as this

	index = 0

	client := pb.NewTransferFileClient(conn)

	for index < len(fileName) {
		sendFile(client)
		index++
	}
	fmt.Println("all file transfer completed")
	conn.Close()
}

func EnumFile(host string, port string, regexp string) {
	address := host + ":" + port
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to server %v, %v", address, err)
	}

	client := pb.NewTransferFileClient(conn)

	enumerateFiles(client, regexp)

	conn.Close()
}

func GetPVC(host string, port string, fileList []string) {
	address := host + ":" + port
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to server %v, %v", address, err)
	}


	client := pb.NewTransferFileClient(conn)

	getPVC(client, fileList)

	conn.Close()
}

