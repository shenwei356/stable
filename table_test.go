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
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	tbl := New().AlignLeft()

	tbl.SetHeader([]string{
		"id",
		"name",
		"sentence",
	})
	tbl.AddRow([]interface{}{1, "Wei Shen", "How are you?"})
	tbl.AddRow([]interface{}{2, "Fake Name", "Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse."})
	tbl.AddRow([]interface{}{3, "Tic Tac", "Doing great!"})

	for _, style := range []*TableStyle{
		StylePlain,
		StyleSimple,
		StyleGrid,
		StyleLight,
		StyleBold,
		StyleDouble,
	} {
		fmt.Printf("style: %s\n%s\n", style.Name, tbl.Render(style))
	}
}
