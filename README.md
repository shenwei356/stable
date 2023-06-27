# stable - streaming pretty text table

[![Go Reference](https://pkg.go.dev/badge/github.com/shenwei356/stable.svg)](https://pkg.go.dev/github.com/shenwei356/stable)

* [Features](#features)
* [Install](#install)
* [Examples](#examples)
* [Styles](#styles)
* [Support](#support)
* [License](#license)
* [Alternate packages](#alternate-packages)

## Features

- **Supporting streaming output**.

  A newly added row is formatted and written to the configured writer immediately.
  It is memory-effective for a large number of rows.
  And it is helpful to pipe the data in shell.

- **Supporting wrapping text or clipping text**.

  The minimum and maximum width of the column can be configured for each column or globally.

- **Configured table styles**.

  Some preset styles are also provided.

## Install

    go get -u github.com/shenwei356/table

## Examples

<p style="color:Tomato;">Note that the output is well-formatted in the terminal.
However, rows containing Unicode are not displayed appropriately in text editors.</p>

1. Basic usages.

        tbl := New().AlignLeft().HumanizeNumbers().MaxWidth(20) //.ClipCell("...")

        tbl.Header([]string{
            "number",
            "name",
            "sentence",
        })
        tbl.AddRow([]interface{}{100, "Wei Shen", "How are you?"})
        tbl.AddRow([]interface{}{1000.1, "沈 伟", "I'm fine, thank you. And you?"})
        tbl.AddRow([]interface{}{100000, "沈伟", "谢谢，我很好，你呢？"})

        fmt.Printf("style: %s\n%s\n", StyleGrid.Name, tbl.Render(StyleGrid))

        style: grid
        +---------+----------+----------------------+
        | number  | name     | sentence             |
        +=========+==========+======================+
        | 100     | Wei Shen | How are you?         |
        +---------+----------+----------------------+
        | 1,000.1 | 沈 伟    | I'm fine, thank      |
        |         |          | you. And you?        |
        +---------+----------+----------------------+
        | 100,000 | 沈伟     | 谢谢，我很好         |
        |         |          | ，你呢？             |
        +---------+----------+----------------------+

        // clipping text instead of wrapping

        fmt.Printf("style: %s\n%s\n", StyleGrid.Name, tbl.ClipCell("...").Render(StyleGrid))

        style: grid
        +---------+----------+----------------------+
        | number  | name     | sentence             |
        +=========+==========+======================+
        | 100     | Wei Shen | How are you?         |
        +---------+----------+----------------------+
        | 1,000.1 | 沈 伟    | I'm fine, thank y... |
        +---------+----------+----------------------+
        | 100,000 | 沈伟     | 谢谢，我很好，你呢？ |
        +---------+----------+----------------------+

1. Custom columns format.

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

        +------------+------------+----------------------+
        |     number |    name    | sentence             |
        +============+============+======================+
        |        100 |  Wei Shen  | How are you?         |
        +------------+------------+----------------------+
        |    1,000.1 |   沈 伟    | I'm fine, thank      |
        |            |            | you. And you?        |
        +------------+------------+----------------------+
        |    100,000 |    沈伟    | 谢谢，我很好         |
        |            |            | ，你呢？             |
        +------------+------------+----------------------+


1. Streaming the output, i.e., a newly added row is formatted and written to the configured writer immediately.

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


        +--------+----------+--------------+
        | number | name     | sentence     |
        +========+==========+==============+
        | 100    | Wei Shen | How are you? |
        +--------+----------+--------------+
        | 1,000. | 沈 伟    | I'm fine,    |
        | 1      |          | thank you.   |
        |        |          | And you?     |
        +--------+----------+--------------+
        | 100,00 | 沈伟     | 谢谢，我     |
        | 0      |          | 很好，你     |
        |        |          | 呢？         |
        +--------+----------+--------------+


1. Custom delimiter for wrapping text.

        tbl := New()

        tbl.SetHeader([]string{
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

        +-------+------------------+----------------------------------------------------+
        | taxid | name             | complete lineage                                   |
        +=======+==================+====================================================+
        | 9606  | Homo sapiens     | cellular organisms;Eukaryota;Opisthokonta;Metazoa; |
        |       |                  | Eumetazoa;Bilateria;Deuterostomia;Chordata;        |
        |       |                  | Craniata;Vertebrata;Gnathostomata;Teleostomi;      |
        |       |                  | Euteleostomi;Sarcopterygii;Dipnotetrapodomorpha;   |
        |       |                  | Tetrapoda;Amniota;Mammalia;Theria;Eutheria;        |
        |       |                  | Boreoeutheria;Euarchontoglires;Primates;           |
        |       |                  | Haplorrhini;Simiiformes;Catarrhini;Hominoidea;     |
        |       |                  | Hominidae;Homininae;Homo;Homo sapiens              |
        +-------+------------------+----------------------------------------------------+
        | 562   | Escherichia coli | cellular organisms;Bacteria;Pseudomonadota;        |
        |       |                  | Gammaproteobacteria;Enterobacterales;              |
        |       |                  | Enterobacteriaceae;Escherichia;Escherichia coli    |
        +-------+------------------+----------------------------------------------------+


## Styles

<p style="color:Tomato;">Note that the output is well-formatted in the terminal.
However, rows containing Unicode are not displayed appropriately in text editors.</p>

    style: plain
    number    name       sentence
    100       Wei Shen   How are you?
    1,000.1   沈 伟      I'm fine, thank
                        you. And you?
    100,000   沈伟       谢谢，我很好
                        ，你呢？

    style: simple
    -------------------------------------------
    number    name       sentence
    -------------------------------------------
    100       Wei Shen   How are you?
    1,000.1   沈 伟      I'm fine, thank
                        you. And you?
    100,000   沈伟       谢谢，我很好
                        ，你呢？
    -------------------------------------------

    style: grid
    +---------+----------+----------------------+
    | number  | name     | sentence             |
    +=========+==========+======================+
    | 100     | Wei Shen | How are you?         |
    +---------+----------+----------------------+
    | 1,000.1 | 沈 伟    | I'm fine, thank      |
    |         |          | you. And you?        |
    +---------+----------+----------------------+
    | 100,000 | 沈伟     | 谢谢，我很好         |
    |         |          | ，你呢？             |
    +---------+----------+----------------------+

    style: light
    ┌---------┬----------┬----------------------┐
    | number  | name     | sentence             |
    ├=========┼==========┼======================┤
    | 100     | Wei Shen | How are you?         |
    ├---------┼----------┼----------------------┤
    | 1,000.1 | 沈 伟    | I'm fine, thank      |
    |         |          | you. And you?        |
    ├---------┼----------┼----------------------┤
    | 100,000 | 沈伟     | 谢谢，我很好         |
    |         |          | ，你呢？             |
    └---------┴----------┴----------------------┘

    style: bold
    ┏━━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━┓
    ┃ number  ┃ name     ┃ sentence             ┃
    ┣━━━━━━━━━╋━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━┫
    ┃ 100     ┃ Wei Shen ┃ How are you?         ┃
    ┣━━━━━━━━━╋━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━┫
    ┃ 1,000.1 ┃ 沈 伟    ┃ I'm fine, thank      ┃
    ┃         ┃          ┃ you. And you?        ┃
    ┣━━━━━━━━━╋━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━┫
    ┃ 100,000 ┃ 沈伟     ┃ 谢谢，我很好         ┃
    ┃         ┃          ┃ ，你呢？             ┃
    ┗━━━━━━━━━┻━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━┛

    style: double
    ╔═════════╦══════════╦══════════════════════╗
    ║ number  ║ name     ║ sentence             ║
    ╠═════════╬══════════╬══════════════════════╣
    ║ 100     ║ Wei Shen ║ How are you?         ║
    ╠═════════╬══════════╬══════════════════════╣
    ║ 1,000.1 ║ 沈 伟    ║ I'm fine, thank      ║
    ║         ║          ║ you. And you?        ║
    ╠═════════╬══════════╬══════════════════════╣
    ║ 100,000 ║ 沈伟     ║ 谢谢，我很好         ║
    ║         ║          ║ ，你呢？             ║
    ╚═════════╩══════════╩══════════════════════╝


## Support

Please [open an issue](https://github.com/shenwei356/stable/issues) to report bugs,
propose new functions or ask for help.

## License

Copyright (c) 2023, Wei Shen (shenwei356@gmail.com)

[MIT License](https://github.com/shenwei356/stable/blob/master/LICENSE)

## Alternate packages

- [go-prettytable](https://github.com/tatsushid/go-prettytable),
  it does not support wrapping cells and it's not flexible to add rows that the number of columns is dynamic.
- [gotabulate](https://github.com/bndr/gotabulate),
  it supports wrapping cells, but it has to read all data in memory before outputing the result.
  We followed the configuration of table styles from this package.
- [go-pretty](https://github.com/jedib0t/go-pretty),
  it supports wrapping cells, but it has to read all data in memory before outputing the result.
  We used some table styles with minor differences in this package.
