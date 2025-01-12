package rest

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/labstack/echo/v4"
)

type PdfService interface {
	CompressPdf(ctx context.Context, fileName string, file multipart.File) (domain.PdfFile, error)
}

type PdfHandler struct {
	Service PdfService
}

func NewPdfHandler(e *echo.Echo, svc PdfService) {
	handler := &PdfHandler{
		Service: svc,
	}
	e.POST("/process/compress", handler.StartCompress)
}

func (a *PdfHandler) StartCompress(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "Failed to get the file"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseError{Message: "Failed to open the file"})
	}
	defer src.Close()

	mimeType, _ := getMimeType(src)
	if mimeType != "application/pdf" {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "File type invalid"})
	}

	ctx := c.Request().Context()
	compressPdfFile, err := a.Service.CompressPdf(ctx, file.Filename, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseError{Message: "Failed to compress pdf"})
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/pdf")
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+compressPdfFile.Name)
	reader := bytes.NewReader(compressPdfFile.Content)
	return c.Stream(http.StatusOK, "application/pdf", reader)
}

func getMimeType(file io.Reader) (string, error) {
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}
	return http.DetectContentType(buf), nil
}
