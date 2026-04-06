# go-googlesheetswrapper

Small read-only wrapper around the official Google Sheets API for Go.

## Features

- Read-only API only (no mutations)
- List sheet names
- Read one sheet as `[][]string`
- Read all sheets as `map[string][][]string`
- Built-in mock client for tests/CI
- Load mock data from JSON
- `context.Context` support for cancellation/timeouts

## Install

```bash
go get github.com/flopp/go-googlesheetswrapper
```

## Real Google Sheets Client

```go
package main

import (
	"context"
	"log"

	googlesheetswrapper "github.com/flopp/go-googlesheetswrapper"
)

func main() {
	ctx := context.Background()

	client, err := googlesheetswrapper.New("YOUR_API_KEY", "YOUR_SHEET_ID")
	if err != nil {
		log.Fatal(err)
	}

	names, err := client.ListSheets(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("sheets: %v", names)

	rows, err := client.ReadSheet(ctx, names[0])
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("first sheet rows: %d", len(rows))
}
```

## Mock Client

```go
ctx := context.Background()

client := googlesheetswrapper.NewMock(map[string][][]string{
	"Events": {
		{"DATE", "NAME"},
		{"2026-04-06", "Community Run"},
	},
})

rows, err := client.ReadSheet(ctx, "Events")
```

## Mock From JSON

```go
client, err := googlesheetswrapper.NewMockFromJSONFile("testdata/sheets.json")
```

JSON format:

```json
{
  "Sheet1": [["A1", "B1"], ["A2", "B2"]],
  "Sheet2": [["X1"]]
}
```

## Extracting and Validating Headers

You can validate and extract a header row (first row) from a sheet using `ExtractHeader`:

```go
rows, err := client.ReadSheet(ctx, "Events")
if err != nil {
	log.Fatal(err)
}

headerIdx, err := googlesheetswrapper.ExtractHeader(rows, []string{"DATE", "NAME"}, false)
if err != nil {
	log.Fatalf("invalid header: %v", err)
}
// headerIdx["DATE"] == 0, headerIdx["NAME"] == 1
```
