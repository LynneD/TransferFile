package server_routine

import (
	"io"
	"context"
)

type StoreFiler interface {
	StoreFile(ctx context.Context, fileName string, reader io.Reader) (int, error)
	CloseFile(fileName string) (error)
}

