package server_routine

import (
	"io"
	"context"
)

type FileManager interface {
	StoreFile(ctx context.Context, fileName string, reader io.Reader) (int64, error)
}

