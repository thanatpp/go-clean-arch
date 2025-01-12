package rest

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PdfHandler struct {
	Service ArticleService
}

func NewPdfHandler(e *echo.Echo) {
	handler := &ArticleHandler{}
	e.POST("/process/compress", handler.StartCompress)
}

func (a *ArticleHandler) StartCompress(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "Failed to get the file"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseError{Message: "Failed to open the file"})
	}
	defer src.Close()

	tempDir := os.TempDir()
	dstPath := filepath.Join(tempDir, file.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save file")
	}
	defer dst.Close()

	err = api.Optimize(src, dst, &model.Configuration{DecodeAllStreams: true})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseError{Message: "PDF optimization failed: " + err.Error()})
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/pdf")
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=compress.pdf")

	return c.Stream(http.StatusOK, "application/pdf", src)
}
