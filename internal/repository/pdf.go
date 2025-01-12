package repository

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

//go:generate mockery --name PdfCpuApi
type PdfCpuApi interface {
	Optimize(rs io.ReadSeeker, w io.Writer, conf *model.Configuration) error
}

type PdfRepository struct {
	pdfCpuApi PdfCpuApi
}

func NewPdfRepository(pdfCpuApi PdfCpuApi) *PdfRepository {
	return &PdfRepository{
		pdfCpuApi: pdfCpuApi,
	}
}

func (m *PdfRepository) Compress(file multipart.File) ([]byte, error) {
	tempfile, err := os.CreateTemp("", "compress.*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		tempfile.Close()
		os.Remove(tempfile.Name())
	}()

	readSeeker, err := toReadSeeker(file)
	if err != nil {
		return nil, fmt.Errorf("failed to process input file: %w", err)
	}

	if err := m.pdfCpuApi.Optimize(readSeeker, tempfile, nil); err != nil {
		return nil, fmt.Errorf("failed to optimize PDF: %w", err)
	}

	compressedData, err := readTempFile(tempfile)
	if err != nil {
		return nil, fmt.Errorf("failed to read optimized file: %w", err)
	}

	return compressedData, nil
}

func toReadSeeker(file multipart.File) (io.ReadSeeker, error) {
	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, file); err != nil {
		return nil, err
	}
	return bytes.NewReader(buffer.Bytes()), nil
}

func readTempFile(file *os.File) ([]byte, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}
