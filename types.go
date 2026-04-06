/ Package googlesheetswrapper provides a small read-only wrapper around
// Google Sheets for listing sheets and reading full-sheet values as strings.
package googlesheetswrapper

import (
	"context"
	"errors"
)

// ErrSheetNotFound is returned when a requested sheet does not exist.
var ErrSheetNotFound = errors.New("sheet not found")

// Client defines read-only access to a Google Spreadsheet.
type Client interface {
	// ListSheets returns all sheet names in the spreadsheet.
	ListSheets(ctx context.Context) ([]string, error)
	// ReadSheet returns all rows and cells of a single sheet as strings.
	ReadSheet(ctx context.Context, name string) ([][]string, error)
	// ReadAll returns all sheets keyed by sheet name.
	ReadAll(ctx context.Context) (map[string][][]string, error)
}
