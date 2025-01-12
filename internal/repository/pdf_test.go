package repository_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/bxcodec/go-clean-arch/internal/repository"
	"github.com/bxcodec/go-clean-arch/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCompressPdf(t *testing.T) {
	mockPdfCpuApi := new(mocks.PdfCpuApi)
	repo := repository.NewPdfRepository(mockPdfCpuApi)

	t.Run("when compress success should return []byte", func(t *testing.T) {
		mockPdfCpuApi.On("Optimize", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		input, _ := os.Open("../resource/test.pdf")
		defer input.Close()

		_, err := repo.Compress(input)

		assert.NoError(t, err)
	})

	t.Run("when compress failed should be return error", func(t *testing.T) {
		mockPdfCpuApi.On("Optimize", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("Compress Error")).Once()

		input, _ := os.Open("../resource/test.pdf")
		defer input.Close()

		_, err := repo.Compress(input)

		assert.Error(t, err)
	})
}
