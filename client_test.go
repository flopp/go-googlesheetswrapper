package googlesheetswrapper

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/sheets/v4"
)

type fakeSheetsAPI struct {
	spreadsheet *sheets.Spreadsheet
	spreadErr   error
	valuesByRng map[string]*sheets.ValueRange
	valuesErr   map[string]error
}

func (f *fakeSheetsAPI) GetSpreadsheet(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
	if f.spreadErr != nil {
		return nil, f.spreadErr
	}
	return f.spreadsheet, nil
}

func (f *fakeSheetsAPI) GetValues(_ context.Context, _ string, readRange string) (*sheets.ValueRange, error) {
	if err, ok := f.valuesErr[readRange]; ok {
		return nil, err
	}
	if v, ok := f.valuesByRng[readRange]; ok {
		return v, nil
	}
	return &sheets.ValueRange{}, nil
}

func TestGoogleClientListSheets(t *testing.T) {
	client := newWithAPI("sheet-id", &fakeSheetsAPI{
		spreadsheet: &sheets.Spreadsheet{
			Sheets: []*sheets.Sheet{
				{Properties: &sheets.SheetProperties{Title: "Alpha"}},
				{Properties: &sheets.SheetProperties{Title: "Beta"}},
			},
		},
	})

	names, err := client.ListSheets(context.Background())
	if err != nil {
		t.Fatalf("ListSheets returned error: %v", err)
	}
	if len(names) != 2 || names[0] != "Alpha" || names[1] != "Beta" {
		t.Fatalf("unexpected sheet names: %#v", names)
	}
}

func TestGoogleClientReadSheetConvertsValues(t *testing.T) {
	client := newWithAPI("sheet-id", &fakeSheetsAPI{
		valuesByRng: map[string]*sheets.ValueRange{
			"Data": {
				Values: [][]interface{}{{"A", 123, true, nil}},
			},
		},
		valuesErr: map[string]error{},
	})

	rows, err := client.ReadSheet(context.Background(), "Data")
	if err != nil {
		t.Fatalf("ReadSheet returned error: %v", err)
	}
	if len(rows) != 1 || len(rows[0]) != 4 {
		t.Fatalf("unexpected row shape: %#v", rows)
	}
	if rows[0][0] != "A" || rows[0][1] != "123" || rows[0][2] != "true" || rows[0][3] != "" {
		t.Fatalf("unexpected row values: %#v", rows[0])
	}
}

func TestGoogleClientReadSheetMissing(t *testing.T) {
	client := newWithAPI("sheet-id", &fakeSheetsAPI{
		valuesByRng: map[string]*sheets.ValueRange{},
		valuesErr: map[string]error{
			"Missing": &googleapi.Error{Code: 400, Message: "Unable to parse range: Missing"},
		},
	})

	_, err := client.ReadSheet(context.Background(), "Missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrSheetNotFound) {
		t.Fatalf("expected ErrSheetNotFound, got: %v", err)
	}
}

func TestGoogleClientReadAll(t *testing.T) {
	client := newWithAPI("sheet-id", &fakeSheetsAPI{
		spreadsheet: &sheets.Spreadsheet{
			Sheets: []*sheets.Sheet{
				{Properties: &sheets.SheetProperties{Title: "One"}},
				{Properties: &sheets.SheetProperties{Title: "Two"}},
			},
		},
		valuesByRng: map[string]*sheets.ValueRange{
			"One": {Values: [][]interface{}{{"1"}}},
			"Two": {Values: [][]interface{}{{"2"}}},
		},
		valuesErr: map[string]error{},
	})

	all, err := client.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if all["One"][0][0] != "1" || all["Two"][0][0] != "2" {
		t.Fatalf("unexpected all data: %#v", all)
	}
}

func TestNewMockBehavior(t *testing.T) {
	client := NewMock(map[string][][]string{
		"B": {{"2"}},
		"A": {{"1"}},
	})

	names, err := client.ListSheets(context.Background())
	if err != nil {
		t.Fatalf("ListSheets returned error: %v", err)
	}
	if len(names) != 2 || names[0] != "A" || names[1] != "B" {
		t.Fatalf("unexpected names: %#v", names)
	}

	rows, err := client.ReadSheet(context.Background(), "A")
	if err != nil {
		t.Fatalf("ReadSheet returned error: %v", err)
	}
	rows[0][0] = "changed"

	rowsAgain, err := client.ReadSheet(context.Background(), "A")
	if err != nil {
		t.Fatalf("ReadSheet returned error: %v", err)
	}
	if rowsAgain[0][0] != "1" {
		t.Fatalf("expected cloned data, got: %#v", rowsAgain)
	}
}

func TestNewMockFromJSON(t *testing.T) {
	jsonData := `{"S1": [["A", 12, true, null], ["B"]]}`
	client, err := NewMockFromJSON(strings.NewReader(jsonData))
	if err != nil {
		t.Fatalf("NewMockFromJSON returned error: %v", err)
	}

	rows, err := client.ReadSheet(context.Background(), "S1")
	if err != nil {
		t.Fatalf("ReadSheet returned error: %v", err)
	}
	if rows[0][0] != "A" || rows[0][1] != "12" || rows[0][2] != "true" || rows[0][3] != "" {
		t.Fatalf("unexpected converted row: %#v", rows[0])
	}
}

func TestNewMockFromJSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mock.json")
	if err := os.WriteFile(path, []byte(`{"Sheet": [["x"]]}`), 0o600); err != nil {
		t.Fatalf("writing temp json: %v", err)
	}

	client, err := NewMockFromJSONFile(path)
	if err != nil {
		t.Fatalf("NewMockFromJSONFile returned error: %v", err)
	}

	rows, err := client.ReadSheet(context.Background(), "Sheet")
	if err != nil {
		t.Fatalf("ReadSheet returned error: %v", err)
	}
	if len(rows) != 1 || len(rows[0]) != 1 || rows[0][0] != "x" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestExtractHeaderOK(t *testing.T) {
	data := [][]string{{"A", "B", "C"}, {"1", "2", "3"}}

	// exact match
	idx, err := ExtractHeader(data, []string{"A", "B", "C"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx["A"] != 0 || idx["B"] != 1 || idx["C"] != 2 {
		t.Fatalf("wrong indices: %#v", idx)
	}

	// subset of required, extras allowed
	idx, err = ExtractHeader(data, []string{"A", "B"}, true)
	if err != nil {
		t.Fatalf("unexpected error with allowExtra=true: %v", err)
	}
	if idx["C"] != 2 {
		t.Fatalf("extra column missing from returned map: %#v", idx)
	}
}

func TestExtractHeaderEmptySheet(t *testing.T) {
	_, err := ExtractHeader(nil, []string{"A"}, true)
	if !errors.Is(err, ErrEmptySheet) {
		t.Fatalf("expected ErrEmptySheet, got: %v", err)
	}
}

func TestExtractHeaderDuplicate(t *testing.T) {
	data := [][]string{{"A", "B", "A"}}
	_, err := ExtractHeader(data, []string{"A", "B"}, true)
	if err == nil {
		t.Fatal("expected error for duplicate column")
	}
}

func TestExtractHeaderMissingRequired(t *testing.T) {
	data := [][]string{{"A", "B"}}
	_, err := ExtractHeader(data, []string{"A", "B", "C"}, true)
	if err == nil {
		t.Fatal("expected error for missing required column")
	}
}

func TestExtractHeaderExtraForbidden(t *testing.T) {
	data := [][]string{{"A", "B", "C"}}
	_, err := ExtractHeader(data, []string{"A", "B"}, false)
	if err == nil {
		t.Fatal("expected error for unexpected column when allowExtra=false")
	}
}
