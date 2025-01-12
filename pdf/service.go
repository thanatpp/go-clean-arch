package pdf

import (
	"context"
	"mime/multipart"

	"github.com/bxcodec/go-clean-arch/domain"
)

//go:generate mockery --name PdfRepository
type PdfRepository interface {
	Compress(file multipart.File) ([]byte, error)
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
