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
	tbl := New().HumanizeNumbers().MaxWidth(40)

	tbl.Header([]string{
		"id",
		"name",
		"sentence",
	})
	tbl.AddRow([]interface{}{100, "Donec Vitae", "Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse."})
	tbl.AddRow([]interface{}{2000, "Quaerat Voluptatem", "At vero eos et accusamus et iusto odio."})
	tbl.AddRow([]interface{}{250, "with	tab", "<-left cell has one tab."})
	tbl.AddRow([]interface{}{250, "with		tab", "<-left cell has two tabs."})
	tbl.AddRow([]interface{}{3000000, "Aliquam lorem", "Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero."})

	// fmt.Printf("%s\n", tbl.Render(StyleGrid))

	for _, style := range []*TableStyle{
		StylePlain,
		StyleSimple,
		StyleThreeLine,
		StyleGrid,
		StyleLight,
		StyleBold,
		StyleDouble,
	} {
		fmt.Printf("style: %s\n%s\n", style.Name, tbl.Render(style))
	}
}

func TestUnicode(t *testing.T) {
	tbl := New().HumanizeNumbers().MaxWidth(20) //.ClipCell("...")

	tbl.Header([]string{
		"id",
		"name",
		"sentence",
	})
	tbl.AddRow([]interface{}{100, "Wei Shen", "How are you?"})
	tbl.AddRow([]interface{}{1000, "沈 伟", "I'm fine, thank you. And you?"})
	tbl.AddRow([]interface{}{1000, "沈	伟", "There's one tab between the two words"})
	tbl.AddRow([]interface{}{100000, "沈伟", "谢谢，我很好，你呢？"})

	fmt.Printf("%s\n", tbl.Render(StyleGrid))
}

func TestCustomColumns(t *testing.T) {
	tbl := New().MinWidth(10).MaxWidth(30)

	tbl.HeaderWithFormat([]Column{
		{Header: "number", MinWidth: 5, MaxWidth: 20, HumanizeNumbers: true, Align: AlignRight},
		{Header: "name", MinWidth: 14, MaxWidth: 25, Align: AlignCenter},
		{Header: "sentence", MaxWidth: 40, Align: AlignLeft},
	})
	tbl.AddRow([]interface{}{100, "Donec Vitae", "Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse."})
	tbl.AddRow([]interface{}{2000, "Quaerat Voluptatem", "At vero eos et accusamus et iusto odio."})
	tbl.AddRow([]interface{}{3000000, "Aliquam lorem", "Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero."})

	fmt.Printf("%s\n", tbl.Render(StyleGrid))
}

func TestStreaming(t *testing.T) {
	tbl := New().MinWidth(10)

	// write to stdout, and determine the max width according to the first row
	tbl.Writer(os.Stdout, 1)
	tbl.Style(StyleGrid)

	tbl.Header([]string{
		"number",
		"name",
		"sentence",
	})

	// when a new row is added, it writes to stdout immediately.
	tbl.AddRow([]interface{}{100, "Donec Vitae", "Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse."})
	tbl.AddRow([]interface{}{2000, "Quaerat Voluptatem", "At vero eos et accusamus et iusto odio."})
	tbl.AddRow([]interface{}{3000000, "Aliquam lorem", "Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero."})

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
