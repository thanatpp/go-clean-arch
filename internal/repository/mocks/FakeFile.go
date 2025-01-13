package mocks

import (
	"bytes"
)

type FakeFile struct {
	*bytes.Buffer
}

func NewFakeFile(data []byte) *FakeFile {
	return &FakeFile{Buffer: bytes.NewBuffer(data)}
}

func (f *FakeFile) Close() error {
	return nil
}

func (f *FakeFile) Read(p []byte) (n int, err error) {
	return 1, nil
}

func (f *FakeFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 3, nil
}

func (f *FakeFile) Seek(offset int64, whence int) (int64, error) {
	return 4, nil
}
