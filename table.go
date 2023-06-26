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
package table

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	"github.com/mattn/go-runewidth"
)

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
	}
	return "unknown"
}

type Column struct {
	Header string
	Align  Align

	MinWidth int
	MaxWidth int
}

type Table struct {
	rows    [][]string // all rows, or buffered rows of the first bufRows lines
	bufRows int        // the number of rows to determin the max/min width of each column

	columns   []Column // configuration of each column
	nColumns  int      // the number of the header or the first row
	dataAdded bool     // a flag to indicate that some data is added, so calling SetHeader() is not allowed
	hasHeader bool     // a flag to say the table has a header

	// statistics of data in rows
	minWidths     []int // min width of each column
	maxWidths     []int // min width of each column
	widthsChecked bool  // a flag to indicate whether the min/max widths of each column is checked

	// global options set by users
	align           Align
	minWidth        int
	maxWidth        int
	wrapDelimiter   rune
	splitConcat     string
	wrapCell        bool
	clipCell        bool
	humanizeNumbers bool

	// so reused datastructure, for avoiding allocate objects repeatedly
	rotate     [][]string  // just for wrapping a row
	wrappedRow []*[]string // just for wrapping a row
	slice      []string
	wrapLocs   []int // position for wrapping
	poolSlice  *sync.Pool

	style *TableStyle // output style

	writer *io.Writer // writer
}

func New() *Table {
	t := new(Table)
	t.style = StylePlain
	return t
}

// --------------------------------------------------------------------------

func (t *Table) Style(style *TableStyle) *Table {
	t.style = style
	return t
}

var ErrInvalidAlign = fmt.Errorf("table: invalid align value")

func (t *Table) AlignLeft() *Table {
	t.align = AlignLeft
	return t
}
func (t *Table) AlignCenter() *Table {
	t.align = AlignCenter
	return t
}
func (t *Table) AlignRight() *Table {
	t.align = AlignRight
	return t
}

func (t *Table) Align(align Align) (*Table, error) {
	switch align {
	case AlignCenter:
		t.align = AlignCenter
	case AlignLeft:
		t.align = AlignLeft
	case AlignRight:
		t.align = AlignRight
	default:
		return nil, ErrInvalidAlign
	}
	return t, nil
}

func (t *Table) MinWidth(w int) *Table {
	t.minWidth = w
	return t
}

func (t *Table) MaxWidth(w int) *Table {
	t.maxWidth = w
	return t
}

func (t *Table) WrapCell(v bool) *Table {
	t.wrapCell = v
	return t
}

func (t *Table) WrapDelimiter(d rune) *Table {
	t.wrapDelimiter = d
	return t
}

func (t *Table) SplitConcat(s string) *Table {
	t.splitConcat = s
	return t
}

func (t *Table) ClipCell(v bool) *Table {
	t.clipCell = v
	return t
}

func (t *Table) HumanizeNumbers(v bool) *Table {
	t.humanizeNumbers = v
	return t
}

// --------------------------------------------------------------------------
var ErrSetHeaderAfterDataAdded = fmt.Errorf("table: setting header is not allowed after some data being added")

func (t *Table) SetHeader(headers []string) (*Table, error) {
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

func (t *Table) SetHeaderWithFormat(headers []Column) (*Table, error) {
	if t.dataAdded {
		return nil, ErrSetHeaderAfterDataAdded
	}
	t.columns = headers
	t.nColumns = len(headers)
	t.hasHeader = true
	return t, nil
}

var ErrLongRow = fmt.Errorf("table: the added row has too many columns")

func (t *Table) parseRow(row []interface{}) ([]string, error) {
	_row := make([]string, len(row))
	var err error
	var s string
	for i, v := range row {
		s, err = convertToString(v, t.humanizeNumbers)
		if err != nil {
			return nil, err
		}
		_row[i] = s
	}
	return _row, nil
}

func (t *Table) addRow(row []interface{}) error {
	_row, err := t.parseRow(row)
	if err != nil {
		return err
	}

	if t.hasHeader {
		if len(row) > t.nColumns {
			return ErrLongRow
		}
	} else if t.columns == nil { // no header and the t.columns is nil
		t.columns = make([]Column, len(row))
		for i := 0; i < len(row); i++ {
			t.columns[i] = Column{}
		}
		t.nColumns = len(row)
	}

	t.rows = append(t.rows, _row)
	t.dataAdded = true

	return nil
}

func (t *Table) AddRow(row []interface{}) error {
	if t.writer == nil {
		t.addRow(row)
		return nil
	}

	if len(t.rows) < t.bufRows {
		t.addRow(row)
	} else if len(t.rows) == t.bufRows {
		// determin the maxWidth

		// write the header

		// write buffered rows to writer
	} else {
		// _row, err := parseRow

		// write the row to writer
	}
	return nil
}

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
	var w, width int
	var r rune
	if t.wrapLocs == nil {
		t.wrapLocs = make([]int, 0, 8)
	}

	locs := t.wrapLocs // do not need locs = locs[:0] here
	var i, j, ploc, loc int
	var cell string
	for i, cell = range row {
		maxWidth = t.maxWidths[i]
		if len(cell) <= maxWidth {
			t.rotate[i] = append(t.rotate[i], cell)
			continue
		}

		width = 0 // accumulative width
		ploc = 0  // position of the previous space
		locs = locs[:0]
		for j, r = range cell {
			w = runewidth.RuneWidth(r)

			if r == '\n' { // force wrapping
				locs = append(locs, j)
				width = 0
				ploc = 0
				continue
			}

			if r == t.wrapDelimiter {
				if width+w >= maxWidth {
					locs = append(locs, ploc)
					width = w
					ploc = 0
				} else {
					ploc = j // position of the previous space
					width += w
				}
				continue
			}

			width += w
		}
		if width >= maxWidth {
			locs = append(locs, ploc)
		}

		ploc = -1
		for _, loc = range locs {
			t.rotate[i] = append(t.rotate[i], cell[ploc+1:loc])
			ploc = loc
		}
		t.rotate[i] = append(t.rotate[i], cell[ploc+1:])
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

func (t *Table) formatCell(text string, width int, align Align) string {
	a := align
	if t.align > 0 { // global align
		a = align
	}

	// here, width need to be >= len(text)
	switch a {
	case AlignCenter:
		n := (width - len(text)) / 2
		return strings.Repeat(" ", n) + text + strings.Repeat(" ", width-len(text)-n)
	case AlignLeft:
		return text + strings.Repeat(" ", width-len(text))
	case AlignRight:
		return strings.Repeat(" ", width-len(text)) + text
	}
	return text + strings.Repeat(" ", width-len(text))
}

func (t *Table) Render(style *TableStyle) []byte {
	if style == nil { // the argument not given
		style = t.style
	}
	if style == nil { // not defined in the object
		style = StyleGrid
	}

	var buf bytes.Buffer
	if t.slice == nil {
		t.slice = make([]string, t.nColumns)
	}
	slice := t.slice
	// determin the maxWidth
	t.checkWidths()

	lenPad2 := len(style.Padding) * 2
	var wrapped bool

	// write the top line
	if style.LineTop.Visible() {
		buf.WriteString(style.LineTop.begin)
		for i, M := range t.maxWidths {
			slice[i] = strings.Repeat(style.LineTop.hline, M+lenPad2)
		}
		buf.WriteString(strings.Join(slice, style.LineTop.sep))
		buf.WriteString(style.LineTop.end)
		buf.WriteString("\n")
	}

	// write the header
	if t.hasHeader {
		buf.WriteString(style.HeaderRow.begin)
		for i, M := range t.maxWidths {
			slice[i] = style.Padding + t.formatCell(t.columns[i].Header, M, t.columns[i].Align) + style.Padding
		}
		buf.WriteString(strings.Join(slice, style.HeaderRow.sep))
		buf.WriteString(style.HeaderRow.end)
		buf.WriteString("\n")

		// line belowHeader
		if style.LineBelowHeader.Visible() {
			buf.WriteString(style.LineBelowHeader.begin)
			for i, M := range t.maxWidths {
				slice[i] = strings.Repeat(style.LineBelowHeader.hline, M+lenPad2)
			}
			buf.WriteString(strings.Join(slice, style.LineBelowHeader.sep))
			buf.WriteString(style.LineBelowHeader.end)
			buf.WriteString("\n")
		}
	}

	// write the row to writer
	jLastLine := len(t.rows) - 1
	hasLineBetweenRows := style.LineBetweenRows.Visible()
	var row2 *[]string
	for j, row := range t.rows {
		// data row
		wrapped = t.formatRow(row)
		if wrapped {
			for _, row2 = range t.wrappedRow {
				buf.WriteString(style.DataRow.begin)
				for i, M := range t.maxWidths {
					slice[i] = style.Padding + t.formatCell((*row2)[i], M, t.columns[i].Align) + style.Padding
				}
				buf.WriteString(strings.Join(slice, style.DataRow.sep))
				buf.WriteString(style.DataRow.end)
				buf.WriteString("\n")

				t.poolSlice.Put(row2)
			}
		} else {
			buf.WriteString(style.DataRow.begin)
			for i, M := range t.maxWidths {
				slice[i] = style.Padding + t.formatCell(row[i], M, t.columns[i].Align) + style.Padding
			}
			buf.WriteString(strings.Join(slice, style.DataRow.sep))
			buf.WriteString(style.DataRow.end)
			buf.WriteString("\n")
		}

		// line between rows
		if hasLineBetweenRows && j < jLastLine {
			buf.WriteString(style.LineBetweenRows.begin)
			for i, M := range t.maxWidths {
				slice[i] = strings.Repeat(style.LineBetweenRows.hline, M+lenPad2)
			}
			buf.WriteString(strings.Join(slice, style.LineBetweenRows.sep))
			buf.WriteString(style.LineBetweenRows.end)
			buf.WriteString("\n")
		}
	}

	// bottom line
	if style.LineBottom.Visible() {
		buf.WriteString(style.LineBottom.begin)
		for i, M := range t.maxWidths {
			slice[i] = strings.Repeat(style.LineBottom.hline, M+lenPad2)
		}
		buf.WriteString(strings.Join(slice, style.LineBottom.sep))
		buf.WriteString(style.LineBottom.end)
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

var ErrNoDataAdded = fmt.Errorf("table: no data added")

func (t *Table) checkWidths() error {
	if t.hasHeader && !t.dataAdded {
		return ErrNoDataAdded
	}

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

		if c.MinWidth > 0 && c.MinWidth > t.minWidths[i] { // use user defined threshold
			t.minWidths[i] = c.MinWidth
		}
		if t.minWidth > 0 && t.minWidth > t.minWidths[i] { // use user defined global threshold
			t.minWidths[i] = t.minWidth
		}
	}

	t.widthsChecked = true

	return nil
}

// --------------------------------------------------------------------------

var ErrWriterRepeatedlySet = fmt.Errorf("table: writer repeatedly set")

// SetWriter sets a writer for render the table, the first bufRows rows will
// be used to determin the maximum width for each cell if they are not defined
// with MaxWidth().
func (t *Table) SetWriter(w *io.Writer, bufRows int) error {
	if t.writer != nil {
		return ErrWriterRepeatedlySet
	}
	t.writer = w
	t.rows = make([][]string, 0, bufRows)
	t.bufRows = bufRows

	return nil
}

func (t *Table) Flush() error {
	// write the bottom line

	t.Flush()
	return nil
}
