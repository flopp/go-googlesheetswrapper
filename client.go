package googlesheetswrapper

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type sheetsAPI interface {
	GetSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error)
	GetValues(ctx context.Context, spreadsheetID, readRange string) (*sheets.ValueRange, error)
}

type googleSheetsAPI struct {
	service *sheets.Service
}

func (g *googleSheetsAPI) GetSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	return g.service.Spreadsheets.Get(spreadsheetID).
		Fields("sheets(properties(title))").
		Context(ctx).
		Do()
}

func (g *googleSheetsAPI) GetValues(ctx context.Context, spreadsheetID, readRange string) (*sheets.ValueRange, error) {
	return g.service.Spreadsheets.Values.Get(spreadsheetID, readRange).
		Context(ctx).
		Do()
}

type googleClient struct {
	sheetID string
	api     sheetsAPI
}

// New creates a read-only Google Sheets client based on API key and spreadsheet ID.
func New(apiKey, sheetID string) (Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("api key is required")
	}
	if strings.TrimSpace(sheetID) == "" {
		return nil, errors.New("sheet ID is required")
	}

	srv, err := sheets.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}

	return newWithAPI(sheetID, &googleSheetsAPI{service: srv}), nil
}

func newWithAPI(sheetID string, api sheetsAPI) Client {
	return &googleClient{sheetID: sheetID, api: api}
}

func (c *googleClient) ListSheets(ctx context.Context) ([]string, error) {
	resp, err := c.api.GetSpreadsheet(ctx, c.sheetID)
	if err != nil {
		return nil, fmt.Errorf("listing sheets: %w", err)
	}

	names := make([]string, 0, len(resp.Sheets))
	for _, sheet := range resp.Sheets {
		if sheet == nil || sheet.Properties == nil {
			continue
		}
		names = append(names, sheet.Properties.Title)
	}

	return names, nil
}

func (c *googleClient) ReadSheet(ctx context.Context, name string) ([][]string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("sheet name is required")
	}

	resp, err := c.api.GetValues(ctx, c.sheetID, name)
	if err != nil {
		if isMissingSheetError(err) {
			return nil, fmt.Errorf("%w: %s", ErrSheetNotFound, name)
		}
		return nil, fmt.Errorf("reading sheet %q: %w", name, err)
	}

	return valueRowsToStrings(resp.Values), nil
}

func (c *googleClient) ReadAll(ctx context.Context) (map[string][][]string, error) {
	names, err := c.ListSheets(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string][][]string, len(names))
	for _, name := range names {
		rows, readErr := c.ReadSheet(ctx, name)
		if readErr != nil {
			return nil, readErr
		}
		result[name] = rows
	}

	return result, nil
}

func valueRowsToStrings(values [][]interface{}) [][]string {
	rows := make([][]string, 0, len(values))
	for _, row := range values {
		cells := make([]string, len(row))
		for i, value := range row {
			if value == nil {
				cells[i] = ""
				continue
			}
			cells[i] = fmt.Sprintf("%v", value)
		}
		rows = append(rows, cells)
	}
	return rows
}

func isMissingSheetError(err error) bool {
	var gErr *googleapi.Error
	if !errors.As(err, &gErr) {
		return false
	}

	msg := strings.ToLower(gErr.Message)
	return strings.Contains(msg, "unable to parse range") || strings.Contains(msg, "unknown range")
}
