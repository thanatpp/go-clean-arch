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
