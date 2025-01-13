package repository

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

//go:generate mockery --name PdfCpuApi
type PdfCpuApi interface {
	Optimize(rs io.ReadSeeker, w io.Writer, conf *model.Configuration) error
	PageCount(rs io.ReadSeeker, conf *model.Configuration) (int, error)
	Split(rs io.ReadSeeker, outDir, fileName string, span int, conf *model.Configuration) error
	SplitByPageNr(rs io.ReadSeeker, outDir, fileName string, pageNrs []int, conf *model.Configuration) error
	MergeCreateFile(inFiles []string, outFile string, dividerPage bool, conf *model.Configuration) (err error)
}

type FileHelper interface {
	Open(name string) (*os.File, error)
}

type PdfRepository struct {
	pdfCpuApi  PdfCpuApi
	fileHelper FileHelper
}

func NewPdfRepository(pdfCpuApi PdfCpuApi, fileHelper FileHelper) *PdfRepository {
	return &PdfRepository{
		pdfCpuApi:  pdfCpuApi,
		fileHelper: fileHelper,
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

func (m *PdfRepository) Split(file multipart.File, pages []int) ([]byte, error) {
	readSeeker, err := toReadSeeker(file)
	if err != nil {
		return nil, err
	}

	name := uuid.NewString()
	if err := m.pdfCpuApi.Split(readSeeker, os.TempDir(), name+".pdf", 1, nil); err != nil {
		return nil, err
	}

	inFiles := make([]string, 0)
	for _, page := range pages {
		inFiles = append(inFiles, filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d%s", name, page, ".pdf")))
	}

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s%s", name, ".pdf"))
	if err := m.pdfCpuApi.MergeCreateFile(inFiles, outputPath, false, nil); err != nil {
		return nil, fmt.Errorf("failed to merge pdf: %w", err)
	}

	output, err := m.fileHelper.Open(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}

	splitData, err := readTempFile(output)
	if err != nil {
		return nil, err
	}

	return splitData, nil
}

func (m *PdfRepository) PageCount(file multipart.File) (int, error) {
	readSeeker, err := toReadSeeker(file)
	if err != nil {
		return 0, fmt.Errorf("failed to process input file: %w", err)
	}
	return m.pdfCpuApi.PageCount(readSeeker, nil)
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
