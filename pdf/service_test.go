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

func TestSplitPdfByRanges(t *testing.T) {
	mockPdfRepo := new(mocks.PdfRepository)
	service := pdf.NewService(mockPdfRepo)

	t.Run("when split success should be return pdfFile", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Split", mock.Anything, mock.Anything).Return([]byte{1, 2}, nil).Once()

		actual, err := service.SplitPdfByRanges(context.TODO(), "test.pdf", input, []int{1, 2})

		assert.NoError(t, err)
		assert.Equal(t, "split_test.pdf", actual.Name)
		assert.Equal(t, []byte{1, 2}, actual.Content)
	})

	t.Run("when split failed should be return error", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Split", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Split Failed")).Once()

		_, err := service.SplitPdfByRanges(context.TODO(), "test.pdf", input, []int{1, 2})

		assert.Error(t, err)
	})
}

func TestRemovePagesPdf(t *testing.T) {
	mockPdfRepo := new(mocks.PdfRepository)
	service := pdf.NewService(mockPdfRepo)

	t.Run("when split success should be return pdfFile", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Split", mock.Anything, mock.Anything).Return([]byte{1, 2}, nil).Once()

		actual, err := service.RemovePagesPdf(context.TODO(), "test.pdf", input, []int{1, 2}, 10)

		assert.NoError(t, err)
		assert.Equal(t, "split_test.pdf", actual.Name)
		assert.Equal(t, []byte{1, 2}, actual.Content)
	})

	t.Run("when split failed should be return error", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Split", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Split Failed")).Once()

		_, err := service.RemovePagesPdf(context.TODO(), "test.pdf", input, []int{1, 2}, 12)

		assert.Error(t, err)
	})
}

func TestSplitAndZipPdfByFixedRange(t *testing.T) {
	mockPdfRepo := new(mocks.PdfRepository)
	service := pdf.NewService(mockPdfRepo)

	t.Run("when split success with result multiple pdf file should return zip file", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Split", mock.Anything, mock.Anything).Return([]byte{1, 2}, nil).Times(2)

		actual, err := service.SplitAndZipPdfByFixedRange(context.TODO(), "zip.pdf", input, [][]int{{1, 2}, {3, 4}})

		assert.NoError(t, err)
		assert.Equal(t, "split_zip.pdf.zip", actual.Name)
		assert.NotNil(t, actual)
	})

	t.Run("when split success with result single pdf file should return pdf file", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Split", mock.Anything, mock.Anything).Return([]byte{1, 2}, nil).Once()

		actual, err := service.SplitAndZipPdfByFixedRange(context.TODO(), "zip.pdf", input, [][]int{{1, 2}})

		assert.NoError(t, err)
		assert.Equal(t, "split_zip.pdf", actual.Name)
		assert.NotNil(t, actual)
	})

	t.Run("when split failed with result multiple pdf file should error", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Split", mock.Anything, mock.Anything).Return([]byte{1, 2}, fmt.Errorf("Error Split")).Once()

		_, err := service.SplitAndZipPdfByFixedRange(context.TODO(), "zip.pdf", input, [][]int{{1, 2}, {3, 4}})

		assert.Error(t, err)
	})

	t.Run("when split failed with result single pdf file should return error", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("Split", mock.Anything, mock.Anything).Return([]byte{1, 2}, fmt.Errorf("Error split")).Once()

		_, err := service.SplitAndZipPdfByFixedRange(context.TODO(), "zip.pdf", input, [][]int{{1, 2}})

		assert.Error(t, err)
	})
}

func TestPageCount(t *testing.T) {
	mockPdfRepo := new(mocks.PdfRepository)
	service := pdf.NewService(mockPdfRepo)

	t.Run("when get pageCount success should be return number of page count", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("PageCount", mock.Anything).Return(12, nil).Once()

		actual, err := service.PageCount(context.TODO(), input)

		assert.NoError(t, err)
		assert.Equal(t, 12, actual)

	})

	t.Run("when get pageCount failed should be return error", func(t *testing.T) {
		input, _ := os.Open("./resource/test.pdf")
		defer input.Close()

		mockPdfRepo.On("PageCount", mock.Anything).Return(0, fmt.Errorf("Page Count Error")).Once()

		_, err := service.PageCount(context.TODO(), input)

		assert.Error(t, err)
	})
}
