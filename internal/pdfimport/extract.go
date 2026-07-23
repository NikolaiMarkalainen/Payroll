package pdfimport

import (
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractTokens reads all non-empty text tokens from a PDF (page order).
func ExtractTokens(path string) ([]string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("pdf open: %w", err)
	}
	defer f.Close()

	var out []string
	for page := 1; page <= r.NumPage(); page++ {
		p := r.Page(page)
		if p.V.IsNull() {
			continue
		}
		rows, err := p.GetTextByRow()
		if err != nil {
			return nil, fmt.Errorf("pdf page %d: %w", page, err)
		}
		for _, row := range rows {
			for _, w := range row.Content {
				t := strings.TrimSpace(w.S)
				if t != "" {
					out = append(out, t)
				}
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("pdf ei sisältänyt tekstiä")
	}
	return out, nil
}
