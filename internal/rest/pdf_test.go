package rest_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/bxcodec/go-clean-arch/internal/rest"
	"github.com/bxcodec/go-clean-arch/internal/rest/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStartCompress(t *testing.T) {
	mockPdfSvc := new(mocks.PdfService)

	createMultipartForm := func(filePath string) (*bytes.Buffer, string, error) {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to open test file: %v", err)
		}
		defer file.Close()

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", filePath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create form file: %v", err)
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, "", fmt.Errorf("failed to copy file to multipart form: %v", err)
		}

		writer.Close()
		return &body, writer.FormDataContentType(), nil
	}

	t.Run("when start compress success should return status 200", func(t *testing.T) {
		testFile := "../resource/test.pdf"
		body, contentType, err := createMultipartForm(testFile)
		if err != nil {
			t.Fatalf("Error creating multipart form: %v", err)
		}

		mockPdfSvc.On("CompressPdf", mock.Anything, mock.Anything, mock.Anything).Return(domain.PdfFile{
			Name:    "compress_test.pdf",
			Content: []byte{1},
		}, nil).Once()

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/process/compress", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		handler := rest.PdfHandler{
			Service: mockPdfSvc,
		}

		err = handler.StartCompress(c)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "pdf")
		assert.Contains(t, rec.Header().Get(echo.HeaderContentDisposition), "compress_test.pdf")

	})

	t.Run("when compress fails should return status 500", func(t *testing.T) {
		testFile := "../resource/test.pdf"
		body, contentType, err := createMultipartForm(testFile)
		if err != nil {
			t.Fatalf("Error creating multipart form: %v", err)
		}

		mockPdfSvc.On("CompressPdf", mock.Anything, mock.Anything, mock.Anything).Return(domain.PdfFile{}, fmt.Errorf("Compress Error")).Once()

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/process/compress", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		handler := rest.PdfHandler{
			Service: mockPdfSvc,
		}

		err = handler.StartCompress(c)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
