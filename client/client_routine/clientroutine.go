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
)


var filename string //etwd Join
//var remotefilename string
var md5array []string
var length int


func sendFile(client pb.TransferFileClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancel()

	stream, err := client.SendFile(ctx)
	if err != nil {
		log.Fatalf("%v.SendFile(_) = _, %v", client, err)
	}

	f, err := os.Open(filename)
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
		sendFileRequest := &pb.SendFileRequest{FileName: filename, Md5Sum:md5str, Data:buf[:n]}
		if err := stream.Send(sendFileRequest); err != nil {
			errorCount++
			log.Fatalf("%v.Send(%v) = %v", stream, sendFileRequest, err)
		} else {
			errorCount = 0
		}
	}

	// close stream and receive result from server
	sendFileResponse, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error 5v, want %v", stream, err, nil)
	}
	log.Printf("Reply from server: %d bytes write", sendFileResponse.BytesWritten)
	fmt.Println(reflect.DeepEqual(md5array, sendFileResponse.Md5Sum))
}

func ClientRoutine(chunksize int, host string, port string, fileName string) {
	address := host + ":" + port
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to server %v, %v", address, err)
	}
	defer conn.Close()

	length = chunksize

	filename = fileName

	client := pb.NewTransferFileClient(conn)

	sendFile(client)
}

