package parsers

import (
	"fmt"
	"io"
	"strings"
)

type PrivatBankParser struct{}

func NewPrivatBankParser() *PrivatBankParser {
	return &PrivatBankParser{}
}

func (p *PrivatBankParser) Parse(reader io.Reader, size int64, filename string) ([]ParsedTransaction, error) {
	lowerName := strings.ToLower(filename)

	if strings.HasSuffix(lowerName, ".xls") || strings.HasSuffix(lowerName, ".xlsx") {
		return parsePrivatBankXLS(reader)
	}
	if strings.HasSuffix(lowerName, ".pdf") {
		return parsePrivatBankPDF(reader, size)
	}
	return nil, fmt.Errorf("непідтримуваний формат файлу: %s", filename)
}