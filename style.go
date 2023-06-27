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

// The data structures are similar to these in https://github.com/bndr/gotabulate.
type TableStyle struct {
	Name string

	LineTop         LineStyle
	LineBelowHeader LineStyle
	LineBetweenRows LineStyle
	LineBottom      LineStyle

	HeaderRow RowStyle
	DataRow   RowStyle
	Padding   string
}

type LineStyle struct {
	begin string
	hline string
	sep   string
	end   string
}

func (s LineStyle) Visible() bool {
	if s.begin != "" || s.hline != "" || s.sep != "" || s.end != "" {
		return true
	}
	return false
}

type RowStyle struct {
	begin string
	sep   string
	end   string
}

var StylePlain = &TableStyle{
	Name: "plain",

	HeaderRow: RowStyle{"", "   ", ""},
	DataRow:   RowStyle{"", "   ", ""},
	Padding:   "",
}

var StyleSimple = &TableStyle{
	Name: "simple",

	LineTop:         LineStyle{"", "-", "-", ""},
	LineBelowHeader: LineStyle{"", "-", "-", ""},
	LineBottom:      LineStyle{"", "-", "-", ""},

	HeaderRow: RowStyle{"", " ", ""},
	DataRow:   RowStyle{"", " ", ""},
	Padding:   " ",
}

var StyleGrid = &TableStyle{
	Name: "grid",

	LineTop:         LineStyle{"+", "-", "+", "+"},
	LineBelowHeader: LineStyle{"+", "=", "+", "+"},
	LineBetweenRows: LineStyle{"+", "-", "+", "+"},
	LineBottom:      LineStyle{"+", "-", "+", "+"},

	HeaderRow: RowStyle{"|", "|", "|"},
	DataRow:   RowStyle{"|", "|", "|"},
	Padding:   " ",
}

var StyleLight = &TableStyle{
	Name: "light",

	LineTop:         LineStyle{"┌", "-", "┬", "┐"},
	LineBelowHeader: LineStyle{"├", "=", "┼", "┤"},
	LineBetweenRows: LineStyle{"├", "-", "┼", "┤"},
	LineBottom:      LineStyle{"└", "-", "┴", "┘"},

	HeaderRow: RowStyle{"|", "|", "|"},
	DataRow:   RowStyle{"|", "|", "|"},
	Padding:   " ",
}

var StyleBold = &TableStyle{
	Name: "bold",

	LineTop:         LineStyle{"┏", "━", "┳", "┓"},
	LineBelowHeader: LineStyle{"┣", "━", "╋", "┫"},
	LineBetweenRows: LineStyle{"┣", "━", "╋", "┫"},
	LineBottom:      LineStyle{"┗", "━", "┻", "┛"},

	HeaderRow: RowStyle{"┃", "┃", "┃"},
	DataRow:   RowStyle{"┃", "┃", "┃"},
	Padding:   " ",
}

var StyleDouble = &TableStyle{
	Name: "double",

	LineTop:         LineStyle{"╔", "═", "╦", "╗"},
	LineBelowHeader: LineStyle{"╠", "═", "╬", "╣"},
	LineBetweenRows: LineStyle{"╠", "═", "╬", "╣"},
	LineBottom:      LineStyle{"╚", "═", "╩", "╝"},

	HeaderRow: RowStyle{"║", "║", "║"},
	DataRow:   RowStyle{"║", "║", "║"},
	Padding:   " ",
}
