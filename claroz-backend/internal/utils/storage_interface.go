package utils

import "mime/multipart"

type FileStorageInterface interface {
	SaveFile(file *multipart.FileHeader) (string, error)
	DeleteFile(path string) error
}
