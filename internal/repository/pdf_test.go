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
	input, _ := os.Open("../resource/test.pdf")
	defer input.Close()

	mockPdfCpuApi := new(mocks.PdfCpuApi)
	mockFileHelper := new(mocks.FileHelper)
	repo := repository.NewPdfRepository(mockPdfCpuApi, mockFileHelper)

	t.Run("when compress success should return []byte", func(t *testing.T) {
		mockPdfCpuApi.On("Optimize", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_, err := repo.Compress(input)

		assert.NoError(t, err)
	})

	t.Run("when compress failed should be return error", func(t *testing.T) {
		mockPdfCpuApi.On("Optimize", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("Compress Error")).Once()

		_, err := repo.Compress(input)

		assert.Error(t, err)
	})
}

func TestSplitPdf(t *testing.T) {
	mockInput, _ := os.Open("../resource/test.fake")
	defer mockInput.Close()

	mockPdfCpuApi := new(mocks.PdfCpuApi)
	mockFileHelper := new(mocks.FileHelper)
	repo := repository.NewPdfRepository(mockPdfCpuApi, mockFileHelper)

	t.Run("when split success should return []byte", func(t *testing.T) {
		mockPdfCpuApi.On("Split", mock.Anything, mock.Anything, mock.Anything, 1, mock.Anything).
			Return(nil).Once()
		mockPdfCpuApi.On("MergeCreateFile", mock.Anything, mock.Anything, false, mock.Anything).
			Return(nil).Once()
		mockFileHelper.On("Open", mock.Anything).Return(mockInput, nil).Once()

		pages := []int{1, 2}

		actual, err := repo.Split(mockInput, pages)

		assert.NoError(t, err, "Splitting the PDF should not return an error")
		assert.NotNil(t, actual, "Split data should not be nil")
	})

	t.Run("when split failed should return error", func(t *testing.T) {
		mockPdfCpuApi.On("Split", mock.Anything, mock.Anything, mock.Anything, 1, mock.Anything).
			Return(fmt.Errorf("Split Error")).Once()

		pages := []int{1, 2}

		_, err := repo.Split(mockInput, pages)

		assert.Error(t, err)
	})

	t.Run("when open file failed should return error", func(t *testing.T) {

		mockPdfCpuApi.On("Split", mock.Anything, mock.Anything, mock.Anything, 1, mock.Anything).
			Return(nil).Once()
		mockPdfCpuApi.On("MergeCreateFile", mock.Anything, mock.Anything, false, mock.Anything).
			Return(nil).Once()
		mockFileHelper.On("Open", mock.Anything).Return(nil, fmt.Errorf("Open File Error")).Once()

		pages := []int{1, 2}

		actual, err := repo.Split(mockInput, pages)

		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestPageCount(t *testing.T) {
	mockInput, _ := os.Open("../resource/test_split.pdf")
	defer mockInput.Close()

	mockPdfCpuApi := new(mocks.PdfCpuApi)
	mockFileHelper := new(mocks.FileHelper)
	repo := repository.NewPdfRepository(mockPdfCpuApi, mockFileHelper)

	t.Run("when page count success should be return page number of pdf", func(t *testing.T) {
		mockPdfCpuApi.On("PageCount", mock.Anything, mock.Anything).Return(1, nil).Once()

		actual, err := repo.PageCount(mockInput)

		assert.NoError(t, err)
		assert.Equal(t, 1, actual)
	})

	t.Run("when page count failed should be return error", func(t *testing.T) {
		mockPdfCpuApi.On("PageCount", mock.Anything, mock.Anything).Return(0, fmt.Errorf("Error Page Count")).Once()

		_, err := repo.PageCount(mockInput)

		assert.Error(t, err)
	})
}
