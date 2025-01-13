package domain

type PdfFile struct {
	Name    string
	Content []byte
}

type SplitPdfFile struct {
	SplitMode   string `form:"split_mode" validate:"required"`
	Ranges      string `form:"ranges"`
	FixedRange  int    `form:"fixed_range"`
	RemovePages string `form:"remove_page"`
}
