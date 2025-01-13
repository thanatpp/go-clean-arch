package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/labstack/echo/v4"
)

const (
	SPLIT_MODE_RANGE        = "ranges"
	SPLIT_MODE_FIXED_RANGE  = "fixed_range"
	SPLIT_MODE_REMOVE_PAGED = "remove_pages"
)

type PdfService interface {
	CompressPdf(ctx context.Context, fileName string, file multipart.File) (domain.PdfFile, error)
	SplitPdfByRanges(ctx context.Context, fileName string, file multipart.File, ranges []int) (domain.PdfFile, error)
	SplitAndZipPdfByFixedRange(ctx context.Context, fileName string, file multipart.File, fra [][]int) (domain.PdfFile, error)
	RemovePagesPdf(ctx context.Context, fileName string, file multipart.File, removePages []int, pageCount int) (domain.PdfFile, error)
	PageCount(ctx context.Context, file multipart.File) (int, error)
}

type PdfHandler struct {
	Service PdfService
}

func NewPdfHandler(e *echo.Echo, svc PdfService) {
	handler := &PdfHandler{
		Service: svc,
	}
	e.POST("/process/compress", handler.StartCompress)
	e.POST("/process/split", handler.StartSplit)
}

func (a *PdfHandler) validateAndOpenFile(c echo.Context) (string, multipart.File, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return "", nil, c.JSON(http.StatusBadRequest, ResponseError{Message: "Failed to get the file"})
	}

	src, err := file.Open()
	if err != nil {
		return "", nil, c.JSON(http.StatusInternalServerError, ResponseError{Message: "Failed to open the file"})
	}
	defer src.Close()

	return file.Filename, src, nil
}

// @Summary Compress a PDF file
// @Description This API compresses the provided PDF file and returns the compressed version.
// @Tags PDF
// @Accept multipart/form-data
// @Param file formData file true "PDF file"
// @Success 200 {file} string "Compressed PDF file"
// @Failure 400 {object} ResponseError "File type is invalid"
// @Failure 500 {object} ResponseError "Failed to compress PDF"
// @Router /process/compress [post]
func (a *PdfHandler) StartCompress(c echo.Context) error {
	fileName, src, err := a.validateAndOpenFile(c)
	if err != nil {
		return err
	}
	defer src.Close()

	src.Seek(0, io.SeekStart)
	ctx := c.Request().Context()
	compressPdfFile, err := a.Service.CompressPdf(ctx, fileName, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseError{Message: "Failed to compress pdf"})
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/pdf")
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+compressPdfFile.Name)
	reader := bytes.NewReader(compressPdfFile.Content)
	return c.Stream(http.StatusOK, "application/pdf", reader)
}

// @Summary Split a PDF file
// @Description This API splits the provided PDF file based on the specified split mode and range
// @Tags PDF
// @Accept multipart/form-data
// @Produce application/pdf, application/zip
// @Param file formData file true "PDF file to be split"
// @Param split_mode formData string true "Split mode (e.g., 'ranges', 'fixed_range', 'remove_pages')"
// @Param ranges formData string false "Page ranges when split_mode = 'ranges' (e.g., '1','5','1-5')"
// @Param remove_page formData string false "Remove pages when split_mode = 'remove_pages' (e.g., '1','5','1-5')"
// @Param fixed_range formData int false "Fixed range when split_mode = fixed_range (e.g., '2', '1')"
// @Success 200 {file} string "Split PDF file"
// @Failure 400 {object} ResponseError "Invalid input or file type"
// @Failure 500 {object} ResponseError "Failed to split PDF"
// @Router /process/split [post]
func (a *PdfHandler) StartSplit(c echo.Context) error {
	req := new(domain.SplitPdfFile)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	fileName, src, err := a.validateAndOpenFile(c)
	if err != nil {
		return err
	}
	defer src.Close()

	src.Seek(0, io.SeekStart)
	ctx := c.Request().Context()
	pageCount, err := a.Service.PageCount(ctx, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseError{Message: "Failed to get page count"})
	}

	switch req.SplitMode {
	case SPLIT_MODE_RANGE:
		if req.Ranges == "" {
			return c.JSON(http.StatusBadRequest, ResponseError{Message: "Invalid Range"})
		}

		ranges, err := validateAndParseRanges(req.Ranges)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		}

		if ranges[len(ranges)-1] > pageCount {
			return c.JSON(http.StatusBadRequest, ResponseError{Message: "Ranges exceed page count"})
		}

		src.Seek(0, io.SeekStart)
		compressedFile, err := a.Service.SplitPdfByRanges(ctx, fileName, src, ranges)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ResponseError{Message: "Failed to split PDF"})
		}

		return a.respondWithPdfOrZip(c, compressedFile)
	case SPLIT_MODE_REMOVE_PAGED:
		if req.RemovePages == "" {
			return c.JSON(http.StatusBadRequest, ResponseError{Message: "Invalid Range"})
		}

		ranges, err := validateAndParseRanges(req.RemovePages)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		}

		if ranges[len(ranges)-1] > pageCount {
			return c.JSON(http.StatusBadRequest, ResponseError{Message: "Ranges exceed page count"})
		}

		src.Seek(0, io.SeekStart)
		compressedFile, err := a.Service.RemovePagesPdf(ctx, fileName, src, ranges, pageCount)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ResponseError{Message: "Failed to split PDF"})
		}

		return a.respondWithPdfOrZip(c, compressedFile)

	case SPLIT_MODE_FIXED_RANGE:
		if req.FixedRange <= 0 {
			return c.JSON(http.StatusBadRequest, ResponseError{Message: "Fixed range must be greater than 0"})
		}

		fixedRange := generateFixedRange(pageCount, req.FixedRange)

		src.Seek(0, io.SeekStart)
		compressedFile, err := a.Service.SplitAndZipPdfByFixedRange(ctx, fileName, src, fixedRange)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ResponseError{Message: "Failed to split PDF by fixed range"})
		}

		return a.respondWithPdfOrZip(c, compressedFile)

	default:
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "Invalid Split Mode"})
	}
}

func (a *PdfHandler) respondWithPdfOrZip(c echo.Context, compressedFile domain.PdfFile) error {
	contentType := "application/pdf"
	if isZipFile(compressedFile.Name) {
		contentType = "application/zip"
	}

	c.Response().Header().Set(echo.HeaderContentType, contentType)
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+compressedFile.Name)
	reader := bytes.NewReader(compressedFile.Content)
	return c.Stream(http.StatusOK, contentType, reader)
}

func isZipFile(fileName string) bool {
	return strings.HasSuffix(fileName, ".zip")
}

func generateFixedRange(totalPages int, fixedRange int) [][]int {
	var result [][]int
	for i := 1; i <= totalPages; i += fixedRange {
		end := i + fixedRange - 1
		if end > totalPages {
			end = totalPages
		}
		var chunk []int
		for j := i; j <= end; j++ {
			chunk = append(chunk, j)
		}
		result = append(result, chunk)
	}

	return result
}

func validateAndParseRanges(inputRange string) ([]int, error) {
	var result []int

	parts := strings.Split(inputRange, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid number in range start: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid number in range end: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("range start cannot be greater than range end")
			}

			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else {
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}
			result = append(result, num)
		}
	}

	return result, nil
}
