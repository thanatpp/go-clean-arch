package repository

import (
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PdfCpuApiImpl struct{}

func (p *PdfCpuApiImpl) Optimize(rs io.ReadSeeker, w io.Writer, conf *model.Configuration) error {
	return api.Optimize(rs, w, conf)
}

func (p *PdfCpuApiImpl) Split(rs io.ReadSeeker, outDir, fileName string, span int, conf *model.Configuration) error {
	return api.Split(rs, outDir, fileName, span, conf)
}

func (p *PdfCpuApiImpl) SplitByPageNr(rs io.ReadSeeker, outDir, fileName string, pageNrs []int, conf *model.Configuration) error {
	return api.SplitByPageNr(rs, outDir, fileName, pageNrs, conf)
}

func (p *PdfCpuApiImpl) PageCount(rs io.ReadSeeker, conf *model.Configuration) (int, error) {
	return api.PageCount(rs, conf)
}

func (p *PdfCpuApiImpl) MergeCreateFile(inFiles []string, outFile string, dividerPage bool, conf *model.Configuration) (err error) {
	return api.MergeCreateFile(inFiles, outFile, dividerPage, conf)
}
