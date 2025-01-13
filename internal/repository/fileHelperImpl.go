package repository

import (
	"os"
)

type FileHelperImpl struct{}

func (p *FileHelperImpl) Open(name string) (*os.File, error) {
	return os.Open(name)
}
