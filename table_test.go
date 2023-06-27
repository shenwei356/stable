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
	"fmt"
	"os"
	"testing"
)

func TestBasic(t *testing.T) {
	tbl := New().HumanizeNumbers().MaxWidth(20) //.ClipCell("...")

	tbl.Header([]string{
		"number",
		"name",
		"sentence",
	})
	tbl.AddRow([]interface{}{100, "Wei Shen", "How are you?"})
	tbl.AddRow([]interface{}{1000.1, "沈 伟", "I'm fine, thank you. And you?"})
	tbl.AddRow([]interface{}{100000, "沈伟", "谢谢，我很好，你呢？"})

	// fmt.Printf("style: %s\n%s\n", StyleGrid.Name, tbl.Render(StyleGrid))
	// return

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

func TestCustomColumns(t *testing.T) {
	tbl := New()

	tbl.HeaderWithFormat([]Column{
		{Header: "number", MinWidth: 10, MaxWidth: 15, HumanizeNumbers: true, Align: AlignRight},
		{Header: "name", MinWidth: 10, MaxWidth: 16, Align: AlignCenter},
		{Header: "sentence", MaxWidth: 20, Align: AlignLeft},
	})
	tbl.AddRow([]interface{}{100, "Wei Shen", "How are you?"})
	tbl.AddRow([]interface{}{1000.1, "沈 伟", "I'm fine, thank you. And you?"})
	tbl.AddRow([]interface{}{100000, "沈伟", "谢谢，我很好，你呢？"})

	fmt.Printf("style: %s\n%s\n", StyleGrid.Name, tbl.Render(StyleGrid))
}
func TestStreaming(t *testing.T) {
	tbl := New().AlignLeft().HumanizeNumbers()

	// write to stdout, and determine the max width according to the first row
	tbl.Writer(os.Stdout, 1)
	tbl.Style(StyleGrid)

	tbl.Header([]string{
		"number",
		"name",
		"sentence",
	})

	// when a new row is added, it writes to stdout immediately.
	tbl.AddRow([]interface{}{100, "Wei Shen", "How are you?"})
	tbl.AddRow([]interface{}{1000.1, "沈 伟", "I'm fine, thank you. And you?"})
	tbl.AddRow([]interface{}{100000, "沈伟", "谢谢，我很好，你呢？"})

	// flush the remaining data
	tbl.Flush()
}

func TestTaxonomicLineages(t *testing.T) {
	tbl := New()

	tbl.Header([]string{
		"taxid",
		"name",
		"complete lineage",
	})
	tbl.AddRow([]interface{}{
		9606,
		"Homo sapiens",
		"cellular organisms;Eukaryota;Opisthokonta;Metazoa;Eumetazoa;Bilateria;Deuterostomia;Chordata;Craniata;Vertebrata;Gnathostomata;Teleostomi;Euteleostomi;Sarcopterygii;Dipnotetrapodomorpha;Tetrapoda;Amniota;Mammalia;Theria;Eutheria;Boreoeutheria;Euarchontoglires;Primates;Haplorrhini;Simiiformes;Catarrhini;Hominoidea;Hominidae;Homininae;Homo;Homo sapiens",
	})
	tbl.AddRow([]interface{}{
		562, "Escherichia coli",
		"cellular organisms;Bacteria;Pseudomonadota;Gammaproteobacteria;Enterobacterales;Enterobacteriaceae;Escherichia;Escherichia coli",
	})

	fmt.Printf("%s\n", tbl.WrapDelimiter(';').AlignLeft().MaxWidth(50).Render(StyleGrid))
}
