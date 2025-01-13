package pdf

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"slices"

	"github.com/bxcodec/go-clean-arch/domain"
)

//go:generate mockery --name PdfRepository
type PdfRepository interface {
	Compress(file multipart.File) ([]byte, error)
	Split(file multipart.File, pages []int) ([]byte, error)
	PageCount(file multipart.File) (int, error)
}

type Service struct {
	pdfRepo PdfRepository
}

func NewService(pdfRepo PdfRepository) *Service {
	return &Service{
		pdfRepo: pdfRepo,
	}
}

func (a *Service) CompressPdf(ctx context.Context, fileName string, file multipart.File) (domain.PdfFile, error) {
	compressContent, err := a.pdfRepo.Compress(file)
	if err != nil {
		return domain.PdfFile{}, err
	}

	outputName := "compressed_" + fileName

	return domain.PdfFile{
		Name:    outputName,
		Content: compressContent,
	}, nil
}

func (a *Service) SplitPdfByRanges(ctx context.Context, fileName string, file multipart.File, ranges []int) (domain.PdfFile, error) {
	splitContent, err := a.pdfRepo.Split(file, ranges)
	if err != nil {
		return domain.PdfFile{}, err
	}

	outputName := "split_" + fileName

	return domain.PdfFile{
		Name:    outputName,
		Content: splitContent,
	}, nil
}

func (a *Service) RemovePagesPdf(ctx context.Context, fileName string, file multipart.File, removePages []int, pageCount int) (domain.PdfFile, error) {
	ranges := make([]int, 0)
	for r := range pageCount {
		if !slices.Contains(removePages, r+1) {
			ranges = append(ranges, r+1)
		}
	}

	splitContent, err := a.pdfRepo.Split(file, ranges)
	if err != nil {
		return domain.PdfFile{}, err
	}

	outputName := "split_" + fileName

	return domain.PdfFile{
		Name:    outputName,
		Content: splitContent,
	}, nil
}

func (a *Service) SplitAndZipPdfByFixedRange(ctx context.Context, fileName string, file multipart.File, fra [][]int) (domain.PdfFile, error) {
	if len(fra) == 1 {
		return a.splitPdfWithoutZip(file, fileName, fra[0])
	}
	return a.splitPdfWithZip(file, fileName, fra)
}

func (a *Service) splitPdfWithoutZip(file multipart.File, fileName string, rangeSet []int) (domain.PdfFile, error) {
	splitContent, err := a.pdfRepo.Split(file, rangeSet)
	if err != nil {
		return domain.PdfFile{}, fmt.Errorf("failed to split pdf for range %v: %w", rangeSet, err)
	}

	outputName := "split_" + fileName
	return domain.PdfFile{
		Name:    outputName,
		Content: splitContent,
	}, nil
}

func (a *Service) splitPdfWithZip(file multipart.File, fileName string, fra [][]int) (domain.PdfFile, error) {
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	for i, ra := range fra {
		file.Seek(0, io.SeekStart)
		splitContent, err := a.pdfRepo.Split(file, ra)
		if err != nil {
			return domain.PdfFile{}, fmt.Errorf("failed to split pdf for range %v: %w", ra, err)
		}

		if err := a.addToZip(zipWriter, splitContent, i); err != nil {
			return domain.PdfFile{}, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return domain.PdfFile{}, fmt.Errorf("failed to close zip writer: %w", err)
	}

	outputName := "split_" + fileName + ".zip"
	return domain.PdfFile{
		Name:    outputName,
		Content: zipBuffer.Bytes(),
	}, nil
}

func (a *Service) addToZip(zipWriter *zip.Writer, content []byte, index int) error {
	fileName := fmt.Sprintf("split_part_%d.pdf", index+1)
	fileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	if _, err := fileWriter.Write(content); err != nil {
		return fmt.Errorf("failed to write split content to zip: %w", err)
	}
	return nil
}

func (a *Service) PageCount(ctx context.Context, file multipart.File) (int, error) {
	return a.pdfRepo.PageCount(file)
}
