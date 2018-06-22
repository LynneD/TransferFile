package server_routine


type StoreFiler interface {
	GetPath() []string
	StoreFile(fileName string) (int, error)
}

