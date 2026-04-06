package googlesheetswrapper

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// NewMockFromJSONFile loads sheet data from JSON and returns a mock client.
func NewMockFromJSONFile(path string) (Client, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening mock json file: %w", err)
	}
	defer f.Close()

	return NewMockFromJSON(f)
}

// NewMockFromJSON loads sheet data from JSON and returns a mock client.
//
// Expected format:
//
//	{
//	  "Sheet1": [["A1", "B1"], ["A2", "B2"]],
//	  "Sheet2": [["X1"]]
//	}
func NewMockFromJSON(r io.Reader) (Client, error) {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	raw := map[string][][]interface{}{}
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding mock json: %w", err)
	}

	result := make(map[string][][]string, len(raw))
	for sheetName, rows := range raw {
		converted := make([][]string, len(rows))
		for rIdx, row := range rows {
			convertedRow := make([]string, len(row))
			for cIdx, cell := range row {
				if cell == nil {
					convertedRow[cIdx] = ""
					continue
				}
				convertedRow[cIdx] = fmt.Sprintf("%v", cell)
			}
			converted[rIdx] = convertedRow
		}
		result[sheetName] = converted
	}

	return NewMock(result), nil
}
