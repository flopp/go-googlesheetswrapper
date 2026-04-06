// Package googlesheetswrapper provides a small read-only wrapper around
// Google Sheets for listing sheets and reading full-sheet values as strings.
package googlesheetswrapper

import (
	"context"
	"errors"
	"fmt"
)

// ErrSheetNotFound is returned when a requested sheet does not exist.
var ErrSheetNotFound = errors.New("sheet not found")

// ErrEmptySheet is returned by ExtractHeader when the sheet has no rows.
var ErrEmptySheet = errors.New("sheet is empty")

// Client defines read-only access to a Google Spreadsheet.
type Client interface {
	// ListSheets returns all sheet names in the spreadsheet.
	ListSheets(ctx context.Context) ([]string, error)
	// ReadSheet returns all rows and cells of a single sheet as strings.
	ReadSheet(ctx context.Context, name string) ([][]string, error)
	// ReadAll returns all sheets keyed by sheet name.
	ReadAll(ctx context.Context) (map[string][][]string, error)
}

// ExtractHeader validates the first row of sheetData and returns a map of
// column name to zero-based column index.
//
// It returns an error if:
//   - sheetData is empty (no rows)
//   - the header row contains duplicate values
//   - any entry in required is absent from the header
//   - allowExtra is false and the header contains columns not listed in required
func ExtractHeader(sheetData [][]string, required []string, allowExtra bool) (map[string]int, error) {
	if len(sheetData) == 0 {
		return nil, ErrEmptySheet
	}
	header := sheetData[0]

	// Build index, catching duplicates.
	index := make(map[string]int, len(header))
	for i, col := range header {
		if _, exists := index[col]; exists {
			return nil, fmt.Errorf("duplicate header column %q", col)
		}
		index[col] = i
	}

	// Check all required columns are present.
	for _, col := range required {
		if _, ok := index[col]; !ok {
			return nil, fmt.Errorf("missing required column %q", col)
		}
	}

	// When extras are forbidden, ensure no column outside required exists.
	if !allowExtra {
		requiredSet := make(map[string]struct{}, len(required))
		for _, col := range required {
			requiredSet[col] = struct{}{}
		}
		for _, col := range header {
			if _, ok := requiredSet[col]; !ok {
				return nil, fmt.Errorf("unexpected column %q", col)
			}
		}
	}

	return index, nil
}
