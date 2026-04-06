package googlesheetswrapper

import (
	"context"
	"fmt"
	"sort"
)

type mockClient struct {
	data map[string][][]string
}

// NewMock creates a read-only client backed by static in-memory data.
func NewMock(data map[string][][]string) Client {
	return &mockClient{data: cloneSheetMap(data)}
}

func (m *mockClient) ListSheets(_ context.Context) ([]string, error) {
	names := make([]string, 0, len(m.data))
	for name := range m.data {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func (m *mockClient) ReadSheet(_ context.Context, name string) ([][]string, error) {
	rows, ok := m.data[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSheetNotFound, name)
	}
	return cloneRows(rows), nil
}

func (m *mockClient) ReadAll(_ context.Context) (map[string][][]string, error) {
	return cloneSheetMap(m.data), nil
}

func cloneSheetMap(in map[string][][]string) map[string][][]string {
	out := make(map[string][][]string, len(in))
	for name, rows := range in {
		out[name] = cloneRows(rows)
	}
	return out
}

func cloneRows(rows [][]string) [][]string {
	out := make([][]string, len(rows))
	for i := range rows {
		out[i] = append([]string(nil), rows[i]...)
	}
	return out
}
