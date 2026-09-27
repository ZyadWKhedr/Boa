package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// Alignment defines cell text alignment
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignRight
	AlignCenter
)

// Column defines header and alignment for a table column
type Column struct {
	Title string
	Align Alignment
}

// Table provides high performance, stable-alignment table rendering for dense CLI output.
type Table struct {
	Columns []Column
	Rows    [][]string
	out     io.Writer
}

// NewTable constructs a table with specified column definitions.
func NewTable(cols ...Column) *Table {
	return &Table{
		Columns: cols,
		Rows:    make([][]string, 0),
		out:     os.Stdout,
	}
}

// SetOutput changes the output destination.
func (t *Table) SetOutput(w io.Writer) {
	t.out = w
}

// AddRow adds a row of values to the table.
func (t *Table) AddRow(cells ...string) {
	t.Rows = append(t.Rows, cells)
}

// Render writes the table to the output writer.
func (t *Table) Render() {
	if len(t.Columns) == 0 {
		return
	}

	colWidths := make([]int, len(t.Columns))
	for i, col := range t.Columns {
		colWidths[i] = utf8.RuneCountInString(StripAnsi(col.Title))
	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(colWidths) {
				w := utf8.RuneCountInString(StripAnsi(cell))
				if w > colWidths[i] {
					colWidths[i] = w
				}
			}
		}
	}

	// Render Header
	var headerLine strings.Builder
	for i, col := range t.Columns {
		padded := padCell(col.Title, colWidths[i], col.Align)
		headerLine.WriteString(Bold(Cyan(padded)))
		if i < len(t.Columns)-1 {
			headerLine.WriteString("  ")
		}
	}
	fmt.Fprintln(t.out, headerLine.String())

	// Render Header Divider
	var divLine strings.Builder
	for i, w := range colWidths {
		divLine.WriteString(Dim(strings.Repeat("─", w)))
		if i < len(colWidths)-1 {
			divLine.WriteString("  ")
		}
	}
	fmt.Fprintln(t.out, divLine.String())

	// Render Rows
	for _, row := range t.Rows {
		var rowLine strings.Builder
		for i := 0; i < len(t.Columns); i++ {
			var cell string
			if i < len(row) {
				cell = row[i]
			}
			padded := padCell(cell, colWidths[i], t.Columns[i].Align)
			rowLine.WriteString(padded)
			if i < len(t.Columns)-1 {
				rowLine.WriteString("  ")
			}
		}
		fmt.Fprintln(t.out, rowLine.String())
	}
}

func padCell(text string, targetWidth int, align Alignment) string {
	plain := StripAnsi(text)
	textLen := utf8.RuneCountInString(plain)
	if textLen >= targetWidth {
		return text
	}

	diff := targetWidth - textLen
	switch align {
	case AlignRight:
		return strings.Repeat(" ", diff) + text
	case AlignCenter:
		left := diff / 2
		right := diff - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	default:
		return text + strings.Repeat(" ", diff)
	}
}
