package server_routine

import (
	"io"
	"context"
	"regexp"
)

type FileManager interface {
	StoreFile(ctx context.Context, fileName string, reader io.Reader) (int64, error)
	GetFileList(ctx context.Context, regexp *regexp.Regexp) ([]string, error)
	GetPVCList(ctx context.Context, fileList []string) ([]string, error)
	Clean(isDone bool, fileName string)
}

