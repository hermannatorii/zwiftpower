package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

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
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting NewSpreadsheetWriter: %v", err)
	}
	sw := spreadsheetWriter{
		max_rows: 1,
		max_cols: 'A',
		id:       spreadsheetID,
		srv:      srv,
	}

	return &sw, nil
}

func (sw *spreadsheetWriter) Write(p []byte) (n int, err error) {
	s := string(p)
	ss := strings.Split(s, ",")

	log.Printf("Appending row to spreadsheet for rider %s, length %d, spreadsheet has %d rows", ss[0], len(ss), len(sw.values))
	sw.values = append(sw.values, ss)
	sw.max_rows += 1

	// Ideally we'd work this out based on the data
	sw.max_cols = 'N'
	return
}

func (sw *spreadsheetWriter) Flush() error {
	rangeData := "sheet1!A1:" + string(sw.max_cols) + strconv.Itoa(sw.max_rows)
	log.Printf("Writing data to spreadsheet range %s", rangeData)
	// values := [][]interface{}{{"sample_A1", "sample_B1"}, {"sample_A2", "sample_B2"}, {"sample_A3", "sample_A3"}}
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
	return err
}

func (sw *spreadsheetWriter) Close() error {
	return nil
}
