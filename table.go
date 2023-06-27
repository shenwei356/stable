// Copyright © 2023 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package stable

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Align is the type of text alignment. Actually, there are only 3 values.
type Align int

const (
	AlignLeft Align = iota + 1
	AlignCenter
	AlignRight
)

func (a Align) String() string {
	switch a {
	case AlignCenter:
		return "center"
	case AlignLeft:
		return "left"
	case AlignRight:
		return "right"
	default:
		return "unknown"
	}
}

// Column is the configuration of a column.
type Column struct {
	Header string // column name
	Align  Align  // text align

	MinWidth int // minimum width
	MaxWidth int // maximum width, it will be overrided by the global MaxWidth of the table

	HumanizeNumbers bool // add comma to numbers, for example 1000 -> 1,000
}

// Table is the table struct.
type Table struct {
	rows [][]string // all rows, or buffered rows of the first bufRows lines when writer is set

	columns   []Column // configuration of each column
	nColumns  int      // the number of the header or the first row
	dataAdded bool     // a flag to indicate that some data is added, so calling SetHeader() is not allowed
	hasHeader bool     // a flag to say the table has a header

	// statistics of data in rows
	minWidths     []int // min width of each column
	maxWidths     []int // min width of each column
	widthsChecked bool  // a flag to indicate whether the min/max widths of each column is checked

	// global options set by users
	align           Align  // text alignment
	minWidth        int    // minimum width
	maxWidth        int    // maximum width
	wrapDelimiter   rune   // delimiter for wrapping cells
	clipCell        bool   // clip cell instead of wrapping
	clipMark        string // mark for indicating the cell if clipped
	humanizeNumbers bool   // add comma to numbers, for example 1000 -> 1,000

	// some reused datastructures, for avoiding allocate objects repeatedly
	slice      []string     // for joining cells of each row
	rotate     [][]string   // only for wrapping a row
	wrappedRow []*[]string  // juonlyst for wrapping a row
	poolSlice  *sync.Pool   // objects pool of string slice which size is the number of columns
	buf        bytes.Buffer // a bytes buffer

	style *TableStyle // output style

	// if the writer is set, the first bufRows rows will  be used to determine
	// the maximum width for each cell if they are not defined with MaxWidth().
	writer        io.Writer
	hasWriter     bool
	bufRows       int // the number of rows to determine the max/min width of each column
	bufRowsDumped bool
	flushed       bool
}

// New creates a new Table object.
func New() *Table {
	t := new(Table)
	t.style = StylePlain
	return t
}

// --------------------------------------------------------------------------

// Style sets the output style.
// If you decide to add all rows before rendering, there's no need to call this method.
// If you want to stream the output, please call this method before adding any rows.
func (t *Table) Style(style *TableStyle) *Table {
	t.style = style
	return t
}

// ErrInvalidAlign means a invalid align value is given.
var ErrInvalidAlign = fmt.Errorf("stable: invalid align value")

// AlignLeft sets the global text alignment as Left.
func (t *Table) AlignLeft() *Table {
	t.align = AlignLeft
	return t
}

// AlignCenter sets the global text alignment as Center.
func (t *Table) AlignCenter() *Table {
	t.align = AlignCenter
	return t
}

// AlignRight sets the global text alignment as Right.
func (t *Table) AlignRight() *Table {
	t.align = AlignRight
	return t
}

// Align sets the global text alignment.
// Only three values are allowed: AlignLeft, AlignCenter, AlignRight.
func (t *Table) Align(align Align) (*Table, error) {
	switch align {
	case AlignLeft:
		t.align = AlignLeft
	case AlignCenter:
		t.align = AlignCenter
	case AlignRight:
		t.align = AlignRight
	default:
		return nil, ErrInvalidAlign
	}
	return t, nil
}

// MinWidth sets the global minimum cell width.
func (t *Table) MinWidth(w int) *Table {
	if t.maxWidth > 0 && w > t.maxWidth { // even bigger than t.maxWidth
		t.minWidth = t.maxWidth
	} else {
		t.minWidth = w
	}
	return t
}

// MaxWidth sets the global maximum cell width.
func (t *Table) MaxWidth(w int) *Table {
	if t.minWidth > 0 && w < t.minWidth { // even smaller than t.minWidth
		t.maxWidth = t.minWidth
	} else {
		t.maxWidth = w
	}
	return t
}

// WrapDelimiter sets the delimiter for wrapping cell text.
// The default value is space.
// Note that in streaming mode (after calling SetWriter())
func (t *Table) WrapDelimiter(d rune) *Table {
	if t.hasWriter && t.dataAdded {
		return t
	}
	t.wrapDelimiter = d
	return t
}

// ClipCell sets the mark to indicate the cell is clipped.
func (t *Table) ClipCell(mark string) *Table {
	t.clipCell = true
	t.clipMark = mark
	return t
}

// HumanizeNumbers makes the numbers more readable by adding commas to numbers. E.g., 1000 -> 1,000.
func (t *Table) HumanizeNumbers() *Table {
	t.humanizeNumbers = true
	return t
}

// --------------------------------------------------------------------------
// ErrSetHeaderAfterDataAdded means that setting header is not allowed after some data being added.
var ErrSetHeaderAfterDataAdded = fmt.Errorf("stable: setting header is not allowed after some data being added")

// Header sets column names.
func (t *Table) Header(headers []string) (*Table, error) {
	if t.dataAdded {
		return nil, ErrSetHeaderAfterDataAdded
	}
	t.columns = make([]Column, len(headers))
	for i, h := range headers {
		t.columns[i] = Column{
			Header: h,
		}
	}
	t.nColumns = len(headers)
	t.hasHeader = true
	return t, nil
}

// HeaderWithFormat sets column names and other configuration of the column.
func (t *Table) HeaderWithFormat(headers []Column) (*Table, error) {
	if t.dataAdded {
		return nil, ErrSetHeaderAfterDataAdded
	}
	t.columns = headers
	t.nColumns = len(headers)
	t.hasHeader = true
	return t, nil
}

// ErrUnmatchedColumnNumber means that the column number
// of the newly added row is not matched with that of previous ones.
var ErrUnmatchedColumnNumber = fmt.Errorf("stable: unmatched column number")

// parseRow convert a list of objects to string slice
func (t *Table) parseRow(row []interface{}) ([]string, error) {
	_row := make([]string, len(row))
	var err error
	var s string
	var humanizeNumbers bool
	for i, v := range row {
		if t.humanizeNumbers {
			humanizeNumbers = true
		} else {
			humanizeNumbers = t.columns[i].HumanizeNumbers
		}

		s, err = convertToString(v, humanizeNumbers)
		if err != nil {
			return nil, err
		}
		_row[i] = s
	}
	return _row, nil
}

// checkRow checks a row.
func (t *Table) checkRow(row []interface{}) ([]string, error) {
	if t.hasHeader {
		if len(row) != t.nColumns {
			return nil, ErrUnmatchedColumnNumber
		}
	} else if t.columns == nil { // no header and the t.columns is nil
		t.columns = make([]Column, len(row))
		for i := 0; i < len(row); i++ {
			t.columns[i] = Column{}
		}
		t.nColumns = len(row)
	} else { // no header
		if len(row) != t.nColumns {
			return nil, ErrUnmatchedColumnNumber
		}
	}

	return t.parseRow(row)
}

var ErrAddRowAfterFlush = fmt.Errorf("stable: calling AddRow is not allowed after calling Flush()")

func (t *Table) AddRowStringSlice(row []string) error {
	tmp := make([]interface{}, len(row))
	for i, v := range row {
		tmp[i] = v
	}

	return t.AddRow(tmp)
}

// AddRow adds a row.
func (t *Table) AddRow(row []interface{}) error {
	if t.hasWriter && t.flushed {
		return ErrAddRowAfterFlush
	}

	// just adds it to buffer
	if !t.hasWriter || len(t.rows) < t.bufRows {
		_row, err := t.checkRow(row)
		if err != nil {
			return err
		}
		t.rows = append(t.rows, _row)
		t.dataAdded = true

		return nil
	}

	// ------------------------------------------------

	style := t.style
	if style == nil { // not defined in the object
		style = StyleGrid
	}

	buf := t.buf
	buf.Reset()

	if t.slice == nil {
		t.slice = make([]string, t.nColumns)
	}
	slice := t.slice

	lenPad2 := len(style.Padding) * 2
	var wrapped bool

	var row2 *[]string

	// ------------------------------------------------

	if t.bufRowsDumped {
		// ------------------------------------------------
		// parse and check row
		_row, err := t.checkRow(row)
		if err != nil {
			return err
		}

		// ------------------------------------------------

		// line between rows
		if style.LineBetweenRows.Visible() {
			buf.WriteString(style.LineBetweenRows.Begin)
			for i, M := range t.maxWidths {
				if M < t.minWidths[i] {
					M = t.minWidths[i]
				}
				slice[i] = strings.Repeat(style.LineBetweenRows.Hline, M+lenPad2)
			}
			buf.WriteString(strings.Join(slice, style.LineBetweenRows.Sep))
			buf.WriteString(style.LineBetweenRows.End)
			buf.WriteString("\n")

			t.writer.Write(buf.Bytes())
			buf.Reset()
		}

		// data row
		wrapped = t.formatRow(_row)
		if wrapped {
			for _, row2 = range t.wrappedRow {
				buf.WriteString(style.DataRow.Begin)
				for i, M := range t.maxWidths {
					if M < t.minWidths[i] {
						M = t.minWidths[i]
					}
					slice[i] = style.Padding + t.formatCell((*row2)[i], M, t.columns[i].Align) + style.Padding
				}
				buf.WriteString(strings.Join(slice, style.DataRow.Sep))
				buf.WriteString(style.DataRow.End)
				buf.WriteString("\n")

				t.writer.Write(buf.Bytes())
				buf.Reset()

				t.poolSlice.Put(row2)
			}
		} else {
			buf.WriteString(style.DataRow.Begin)
			for i, M := range t.maxWidths {
				if M < t.minWidths[i] {
					M = t.minWidths[i]
				}
				slice[i] = style.Padding + t.formatCell(_row[i], M, t.columns[i].Align) + style.Padding
			}
			buf.WriteString(strings.Join(slice, style.DataRow.Sep))
			buf.WriteString(style.DataRow.End)
			buf.WriteString("\n")

			t.writer.Write(buf.Bytes())
			buf.Reset()
		}

		return nil
	}

	// ------------------------------------------------

	if len(t.rows) == t.bufRows {
		// determine the minWidth and maxWidth
		t.checkWidths()

		_row, err := t.checkRow(row)
		if err != nil {
			return err
		}
		t.rows = append(t.rows, _row)
		t.dataAdded = true

		// write the top line
		if style.LineTop.Visible() {
			buf.WriteString(style.LineTop.Begin)
			for i, M := range t.maxWidths {
				if M < t.minWidths[i] {
					M = t.minWidths[i]
				}
				slice[i] = strings.Repeat(style.LineTop.Hline, M+lenPad2)
			}
			buf.WriteString(strings.Join(slice, style.LineTop.Sep))
			buf.WriteString(style.LineTop.End)
			buf.WriteString("\n")

			t.writer.Write(buf.Bytes())
			buf.Reset()
		}

		// write the header
		if t.hasHeader {
			_row := make([]string, t.nColumns)
			for i, c := range t.columns {
				_row[i] = c.Header
			}
			wrapped = t.formatRow(_row)
			if wrapped {
				for _, row2 = range t.wrappedRow {
					buf.WriteString(style.HeaderRow.Begin)
					for i, M := range t.maxWidths {
						if M < t.minWidths[i] {
							M = t.minWidths[i]
						}
						slice[i] = style.Padding + t.formatCell((*row2)[i], M, t.columns[i].Align) + style.Padding
					}
					buf.WriteString(strings.Join(slice, style.HeaderRow.Sep))
					buf.WriteString(style.HeaderRow.End)
					buf.WriteString("\n")

					t.writer.Write(buf.Bytes())
					buf.Reset()

					t.poolSlice.Put(row2)
				}
			} else {
				buf.WriteString(style.HeaderRow.Begin)
				for i, M := range t.maxWidths {
					if M < t.minWidths[i] {
						M = t.minWidths[i]
					}
					slice[i] = style.Padding + t.formatCell(_row[i], M, t.columns[i].Align) + style.Padding
				}
				buf.WriteString(strings.Join(slice, style.HeaderRow.Sep))
				buf.WriteString(style.HeaderRow.End)
				buf.WriteString("\n")

				t.writer.Write(buf.Bytes())
				buf.Reset()
			}

			// line belowHeader
			if style.LineBelowHeader.Visible() {
				buf.WriteString(style.LineBelowHeader.Begin)
				for i, M := range t.maxWidths {
					if M < t.minWidths[i] {
						M = t.minWidths[i]
					}
					slice[i] = strings.Repeat(style.LineBelowHeader.Hline, M+lenPad2)
				}
				buf.WriteString(strings.Join(slice, style.LineBelowHeader.Sep))
				buf.WriteString(style.LineBelowHeader.End)
				buf.WriteString("\n")

				t.writer.Write(buf.Bytes())
				buf.Reset()
			}
		}

		// write the rows
		hasLineBetweenRows := style.LineBetweenRows.Visible()
		for j, _row := range t.rows {
			// line between rows
			if hasLineBetweenRows && j > 0 {
				buf.WriteString(style.LineBetweenRows.Begin)
				for i, M := range t.maxWidths {
					if M < t.minWidths[i] {
						M = t.minWidths[i]
					}
					slice[i] = strings.Repeat(style.LineBetweenRows.Hline, M+lenPad2)
				}
				buf.WriteString(strings.Join(slice, style.LineBetweenRows.Sep))
				buf.WriteString(style.LineBetweenRows.End)
				buf.WriteString("\n")

				t.writer.Write(buf.Bytes())
				buf.Reset()
			}

			// data row
			wrapped = t.formatRow(_row)
			if wrapped {
				for _, row2 = range t.wrappedRow {
					buf.WriteString(style.DataRow.Begin)
					for i, M := range t.maxWidths {
						if M < t.minWidths[i] {
							M = t.minWidths[i]
						}
						slice[i] = style.Padding + t.formatCell((*row2)[i], M, t.columns[i].Align) + style.Padding
					}
					buf.WriteString(strings.Join(slice, style.DataRow.Sep))
					buf.WriteString(style.DataRow.End)
					buf.WriteString("\n")

					t.writer.Write(buf.Bytes())
					buf.Reset()

					t.poolSlice.Put(row2)
				}
			} else {
				buf.WriteString(style.DataRow.Begin)
				for i, M := range t.maxWidths {
					if M < t.minWidths[i] {
						M = t.minWidths[i]
					}
					slice[i] = style.Padding + t.formatCell(_row[i], M, t.columns[i].Align) + style.Padding
				}
				buf.WriteString(strings.Join(slice, style.DataRow.Sep))
				buf.WriteString(style.DataRow.End)
				buf.WriteString("\n")

				t.writer.Write(buf.Bytes())
				buf.Reset()
			}
		}

		t.bufRowsDumped = true
	}

	return nil
}

// formatRow wraps or clips cells.
// the returned value indicate if any cells are wrapped
func (t *Table) formatRow(row []string) bool {
	// -------------------------------------------------------------
	// initialize some data structures

	if t.rotate == nil {
		t.rotate = make([][]string, t.nColumns)
		for i := range t.rotate {
			t.rotate[i] = make([]string, 0, 8)
		}
	} else {
		for i := range t.rotate {
			t.rotate[i] = t.rotate[i][:0]
		}
	}

	if t.wrappedRow == nil {
		t.wrappedRow = make([]*[]string, 0, 8)
	} else {
		t.wrappedRow = t.wrappedRow[:0]
	}

	if t.poolSlice == nil {
		t.poolSlice = &sync.Pool{New: func() interface{} {
			tmp := make([]string, t.nColumns)
			return &tmp
		}}
	}

	if t.wrapDelimiter == 0 {
		t.wrapDelimiter = ' '
	}

	// -------------------------------------------------------------

	var needWrap = false
	for i, c := range row {
		if len(c) > t.maxWidths[i] {
			needWrap = true
		}
	}
	if !needWrap {
		return false
	}

	// -------------------------------------------------------------

	var maxWidth int
	var w int
	var r rune

	var i, j int
	var cell string
	var workingLine string
	var spacePos charPos
	var lastPos charPos
	lenClipMark := len(t.clipMark)
	for i, cell = range row {
		maxWidth = t.maxWidths[i]

		if maxWidth < t.minWidth {
			maxWidth = t.minWidth
		}

		if len(cell) <= maxWidth {
			t.rotate[i] = append(t.rotate[i], cell)
			continue
		}

		// ---------------------------------------------------
		// clip

		if t.clipCell && len(cell) > maxWidth {
			if lenClipMark > maxWidth {
				t.clipMark = ""
				lenClipMark = len(t.clipMark)
			}
			t.rotate[i] = append(t.rotate[i], runewidth.Truncate(cell, maxWidth, t.clipMark))
			continue
		}

		// ---------------------------------------------------
		// wrap

		// modify from https://github.com/donatj/wordwrap

		workingLine = ""
		spacePos.pos = 0
		spacePos.size = 0
		lastPos.pos = 0
		lastPos.size = 0

		for _, r = range cell {
			w = utf8.RuneLen(r)

			workingLine += string(r)

			if r == t.wrapDelimiter {
				spacePos.pos = len(workingLine)
				spacePos.size = w
			}

			if len(workingLine) >= maxWidth {
				if spacePos.size > 0 {
					t.rotate[i] = append(t.rotate[i], workingLine[0:spacePos.pos])

					workingLine = workingLine[spacePos.pos:]
				} else {
					if len(workingLine) > maxWidth {
						t.rotate[i] = append(t.rotate[i], workingLine[0:lastPos.pos])
						workingLine = workingLine[lastPos.pos:]
					} else {
						t.rotate[i] = append(t.rotate[i], workingLine)
						workingLine = ""
					}
				}

				if len(t.rotate[i][len(t.rotate[i])-1]) > maxWidth {
					panic("attempted to cut character")
				}

				spacePos.pos = 0
				spacePos.size = 0
			}

			lastPos.pos = len(workingLine)
			lastPos.size = w
		}

		if workingLine != "" {
			t.rotate[i] = append(t.rotate[i], workingLine)
		}
	}

	var maxRow int
	for _, tmp := range t.rotate {
		if len(tmp) > maxRow {
			maxRow = len(tmp)
		}
	}

	var row2 *[]string

	for j = 0; j < maxRow; j++ {
		row2 = t.poolSlice.Get().(*[]string)
		for i = 0; i < t.nColumns; i++ {
			if j+1 > len(t.rotate[i]) {
				(*row2)[i] = ""
			} else {
				(*row2)[i] = t.rotate[i][j]
			}
		}
		t.wrappedRow = append(t.wrappedRow, row2)
	}

	return true
}

type charPos struct {
	pos, size int
}

// formatCell formats a cell with given width and text alignment.
func (t *Table) formatCell(text string, width int, align Align) string {
	a := align
	if t.align > 0 { // global align
		a = t.align
	}

	lenText := runewidth.StringWidth(text)

	// here, width need to be >= len(text)
	if width-lenText < 0 {
		panic("wrapping/clipping method error, please contact the author")
	}

	var out string
	switch a {
	case AlignCenter:
		n := (width - lenText) / 2
		out = strings.Repeat(" ", n) + text + strings.Repeat(" ", width-lenText-n)
	case AlignLeft:
		out = text + strings.Repeat(" ", width-lenText)
	case AlignRight:
		out = strings.Repeat(" ", width-lenText) + text
	default:
		out = text + strings.Repeat(" ", width-lenText)
	}
	return out
}

// Render render all data with give style.
func (t *Table) Render(style *TableStyle) []byte {
	if style == nil { // the argument not given
		style = t.style
	}
	if style == nil { // not defined in the object
		style = StyleGrid
	}

	buf := t.buf
	buf.Reset()

	if t.slice == nil {
		t.slice = make([]string, t.nColumns)
	}
	slice := t.slice

	lenPad2 := len(style.Padding) * 2
	var wrapped bool

	// determine the minWidth and maxWidth
	t.checkWidths()

	// write the top line
	if style.LineTop.Visible() {
		buf.WriteString(style.LineTop.Begin)
		for i, M := range t.maxWidths {
			if M < t.minWidths[i] {
				M = t.minWidths[i]
			}
			slice[i] = strings.Repeat(style.LineTop.Hline, M+lenPad2)
		}
		buf.WriteString(strings.Join(slice, style.LineTop.Sep))
		buf.WriteString(style.LineTop.End)
		buf.WriteString("\n")
	}

	// write the header
	var row2 *[]string
	if t.hasHeader {
		_row := make([]string, t.nColumns)
		for i, c := range t.columns {
			_row[i] = c.Header
		}
		wrapped = t.formatRow(_row)
		if wrapped {
			for _, row2 = range t.wrappedRow {
				buf.WriteString(style.HeaderRow.Begin)
				for i, M := range t.maxWidths {
					if M < t.minWidths[i] {
						M = t.minWidths[i]
					}
					slice[i] = style.Padding + t.formatCell((*row2)[i], M, t.columns[i].Align) + style.Padding
				}
				buf.WriteString(strings.Join(slice, style.HeaderRow.Sep))
				buf.WriteString(style.HeaderRow.End)
				buf.WriteString("\n")

				t.poolSlice.Put(row2)
			}
		} else {
			buf.WriteString(style.HeaderRow.Begin)
			for i, M := range t.maxWidths {
				if M < t.minWidths[i] {
					M = t.minWidths[i]
				}
				slice[i] = style.Padding + t.formatCell(_row[i], M, t.columns[i].Align) + style.Padding
			}
			buf.WriteString(strings.Join(slice, style.HeaderRow.Sep))
			buf.WriteString(style.HeaderRow.End)
			buf.WriteString("\n")
		}

		// line belowHeader
		if style.LineBelowHeader.Visible() {
			buf.WriteString(style.LineBelowHeader.Begin)
			for i, M := range t.maxWidths {
				if M < t.minWidths[i] {
					M = t.minWidths[i]
				}
				slice[i] = strings.Repeat(style.LineBelowHeader.Hline, M+lenPad2)
			}
			buf.WriteString(strings.Join(slice, style.LineBelowHeader.Sep))
			buf.WriteString(style.LineBelowHeader.End)
			buf.WriteString("\n")
		}
	}

	// write the rows
	hasLineBetweenRows := style.LineBetweenRows.Visible()
	for j, _row := range t.rows {
		// line between rows
		if hasLineBetweenRows && j > 0 {
			buf.WriteString(style.LineBetweenRows.Begin)
			for i, M := range t.maxWidths {
				if M < t.minWidths[i] {
					M = t.minWidths[i]
				}
				slice[i] = strings.Repeat(style.LineBetweenRows.Hline, M+lenPad2)
			}
			buf.WriteString(strings.Join(slice, style.LineBetweenRows.Sep))
			buf.WriteString(style.LineBetweenRows.End)
			buf.WriteString("\n")
		}

		// data row
		wrapped = t.formatRow(_row)
		if wrapped {
			for _, row2 = range t.wrappedRow {
				buf.WriteString(style.DataRow.Begin)
				for i, M := range t.maxWidths {
					if M < t.minWidths[i] {
						M = t.minWidths[i]
					}
					slice[i] = style.Padding + t.formatCell((*row2)[i], M, t.columns[i].Align) + style.Padding
				}
				buf.WriteString(strings.Join(slice, style.DataRow.Sep))
				buf.WriteString(style.DataRow.End)
				buf.WriteString("\n")

				t.poolSlice.Put(row2)
			}
		} else {
			buf.WriteString(style.DataRow.Begin)
			for i, M := range t.maxWidths {
				if M < t.minWidths[i] {
					M = t.minWidths[i]
				}
				slice[i] = style.Padding + t.formatCell(_row[i], M, t.columns[i].Align) + style.Padding
			}
			buf.WriteString(strings.Join(slice, style.DataRow.Sep))
			buf.WriteString(style.DataRow.End)
			buf.WriteString("\n")
		}
	}

	// bottom line
	if style.LineBottom.Visible() {
		buf.WriteString(style.LineBottom.Begin)
		for i, M := range t.maxWidths {
			if M < t.minWidths[i] {
				M = t.minWidths[i]
			}
			slice[i] = strings.Repeat(style.LineBottom.Hline, M+lenPad2)
		}
		buf.WriteString(strings.Join(slice, style.LineBottom.Sep))
		buf.WriteString(style.LineBottom.End)
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

// ErrNoDataAdded means not data is added. Not used.
var ErrNoDataAdded = fmt.Errorf("stable: no data added")

// checkWidths determine the minimum and maximum widths of each column.
func (t *Table) checkWidths() error {
	// if t.hasHeader && !t.dataAdded {
	// 	return ErrNoDataAdded
	// }

	t.minWidths = make([]int, t.nColumns)
	for i := range t.minWidths {
		t.minWidths[i] = math.MaxInt
	}
	t.maxWidths = make([]int, t.nColumns)

	var i, l int
	var c Column
	if t.hasHeader {
		for i, c = range t.columns {
			l = len(c.Header)
			if l > t.maxWidths[i] {
				t.maxWidths[i] = l
			}
			if l < t.minWidths[i] {
				t.minWidths[i] = l
			}
		}
	}

	var v string
	for _, row := range t.rows {
		for i, v = range row {
			l = len(v)
			if l > t.maxWidths[i] {
				t.maxWidths[i] = l
			}
			if l < t.minWidths[i] {
				t.minWidths[i] = l
			}
		}
	}

	for i, c := range t.columns {
		if c.MaxWidth > 0 && c.MaxWidth < t.maxWidths[i] { // use user defined threshold
			t.maxWidths[i] = c.MaxWidth
		}
		if t.maxWidth > 0 && t.maxWidth < t.maxWidths[i] { // use user defined global threshold
			t.maxWidths[i] = t.maxWidth
		}

		if t.maxWidths[i] < 5 {
			t.maxWidths[i] = 5
		}

		if c.MinWidth > 0 && c.MinWidth > t.minWidths[i] { // use user defined threshold
			t.minWidths[i] = c.MinWidth
		}
		if t.minWidth > 0 { // use user defined global threshold
			t.minWidths[i] = t.minWidth
		}
	}
	t.widthsChecked = true

	return nil
}

// --------------------------------------------------------------------------

// ErrWriterRepeatedlySet means that the writer is repeatedly set.
var ErrWriterRepeatedlySet = fmt.Errorf("stable: writer repeatedly set")

// Writer sets a writer for render the table. The first bufRows rows will
// be used to determine the maximum width for each cell if they are not defined
// with MaxWidth().
// So a newly added row (Addrow()) is formatted and written to the configured writer immediately.
// It is memory-effective for a large number of rows.
// And it is helpful to pipe the data in shell.
// Do not forget to call Flush() after adding all rows.
func (t *Table) Writer(w io.Writer, bufRows uint) error {
	if t.hasWriter {
		return ErrWriterRepeatedlySet
	}
	t.writer = w
	t.hasWriter = true
	if bufRows < 1 { // can not be 0
		bufRows = 1
	}
	t.rows = make([][]string, 0, bufRows)
	t.bufRows = int(bufRows)

	return nil
}

// Flush dumps the remaining data.
func (t *Table) Flush() {
	t.flushed = true

	style := t.style
	if style == nil { // not defined in the object
		style = StyleGrid
	}

	buf := t.buf
	buf.Reset()

	if t.slice == nil {
		t.slice = make([]string, t.nColumns)
	}
	slice := t.slice

	lenPad2 := len(style.Padding) * 2

	// ------------------------------------------------
	// only need to append the bottown line

	if t.bufRowsDumped {
		// bottom line
		if style.LineBottom.Visible() {
			buf.WriteString(style.LineBottom.Begin)
			for i, M := range t.maxWidths {
				if M < t.minWidths[i] {
					M = t.minWidths[i]
				}
				slice[i] = strings.Repeat(style.LineBottom.Hline, M+lenPad2)
			}
			buf.WriteString(strings.Join(slice, style.LineBottom.Sep))
			buf.WriteString(style.LineBottom.End)
			buf.WriteString("\n")

			t.writer.Write(buf.Bytes())
			buf.Reset()
		}
		return
	}

	// ------------------------------------------------
	// dump all buffered line

	t.writer.Write(t.Render(style))
	buf.Reset()

	return
}
