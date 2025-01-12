package pdf_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bxcodec/go-clean-arch/pdf"
	"github.com/bxcodec/go-clean-arch/pdf/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCompressPdf(t *testing.T) {
	mockPdfRepo := new(mocks.PdfRepository)
	service := pdf.NewService(mockPdfRepo)

	t.Run("when compress success should be return PdfFile", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Compress", mock.Anything).Return([]byte{1}, nil).Once()

		actual, err := service.CompressPdf(context.TODO(), "input.pdf", input)

		assert.NoError(t, err)
		assert.Equal(t, "compressed_input.pdf", actual.Name)
		assert.Equal(t, []byte{1}, actual.Content)
	})

	t.Run("when compress failed should be return error", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Compress", mock.Anything).Return(nil, fmt.Errorf("Compress Error")).Once()

		_, err := service.CompressPdf(context.TODO(), "input.pdf", input)

		assert.Error(t, err)
	})
}
