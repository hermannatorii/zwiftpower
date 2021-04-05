package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"google.golang.org/api/sheets/v4"
)

type spreadsheetWriter struct {
	srv      *sheets.Service
	max_rows int
	max_cols rune
	values   [][]string
	id       string
}

func NewSpreadsheetWriter(ctx context.Context, spreadsheetID string) (*spreadsheetWriter, error) {
	log.Printf("Getting new spreadsheetWriter")
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting NewSpreadsheetWriter: %v", err)
	}

	// Start at row 2 to leave the header row in place
	sw := spreadsheetWriter{
		max_rows: 2,
		max_cols: 'A',
		id:       spreadsheetID,
		srv:      srv,
	}

	return &sw, nil
}

func (sw spreadsheetWriter) Write(p []byte) (n int, err error) {
	log.Printf("spreadsheet writing no-op")
	return len(p), nil
}

func (sw *spreadsheetWriter) WriteRow(record []string) error {
	log.Printf("Appending row to spreadsheet for rider %s, length %d", record[0], len(record))
	sw.values = append(sw.values, record)
	sw.max_rows += 1

	// Ideally we'd work this out based on the data
	sw.max_cols = 'N'
	log.Printf("Spreadsheet data has %d rows", len(sw.values))
	return nil
}

func (sw *spreadsheetWriter) Flush() {
	// Leave the header row intact
	rangeData := fmt.Sprintf("sheet1!A2:%c%d", sw.max_cols, sw.max_rows)
	log.Printf("Writing data to spreadsheet range %s", rangeData)
	values := make([][]interface{}, len(sw.values))
	for i, row := range sw.values {
		v := make([]interface{}, len(row))
		for j, col := range row {
			v[j] = col
		}
		values[i] = v
	}

	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
	}
	rb.Data = append(rb.Data, &sheets.ValueRange{
		Range:  rangeData,
		Values: values,
	})
	_, err := sw.srv.Spreadsheets.Values.BatchUpdate(sw.id, rb).Do()
	if err != nil {
		log.Printf("writing to spreadsheet: %v", err)
	}
	return
}

func (sw *spreadsheetWriter) Close() error {
	return nil
}

type rowWriter interface {
	WriteRow(record []string) error
	Flush()
}

func NewRowWriter(w io.Writer) rowWriter {
	sw, ok := w.(*spreadsheetWriter)
	if ok {
		log.Printf("This is a spreadsheetWriter")
		return sw
	}

	log.Printf("This is a csv.Writer")
	m := &myCSV{
		csv.NewWriter(w),
	}

	return m
}

type myCSV struct {
	*csv.Writer
}

func (m *myCSV) WriteRow(record []string) error {
	return m.Writer.Write(record)
}

// func (m *myCSV) Flush() {

// }

// func (r *rowWriter) Write(record []string) error {
// 	log.Printf("Writing row %v", record)
// 	if r.c != nil {
// 		log.Printf("using csv.WRite")
// 		return r.c.Write(record)
// 	}

// 	log.Printf("using sw.WriteRow")
// 	return r.sw.WriteRow(record)
// }

// func (r *rowWriter) Flush() {
// 	if r.c != nil {
// 		log.Printf("using csv.Flush")
// 		r.c.Flush()
// 	}

// 	log.Printf("using sw.Flush")
// 	r.sw.Flush()
// }
