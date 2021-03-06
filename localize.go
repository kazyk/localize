package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var langs = []string{
	"ja", "en", "th", "es", "fr", "vi", "zh-Hant",
}

type Item struct {
	File         string
	Key          string
	Comment      string
	Localization map[string]string
}

func NewItem(filename string) *Item {
	return &Item{
		File:         filename,
		Localization: map[string]string{},
	}
}

func (i Item) String() string {
	components := []string{i.File, i.Key, i.Comment}
	for _, l := range langs {
		t := fmt.Sprintf("%v = %v", l, i.Localization[l])
		components = append(components, t)
	}
	return strings.Join(components, "\n")
}

func MergeItems(dst map[string]Item, items []Item) {
	for _, i := range items {
		item, ok := dst[i.Key]
		if ok {
			if len(item.Comment) > 0 && len(i.Comment) > 0 && item.Comment != i.Comment {
				fmt.Fprintf(os.Stderr, `warning: different comments found: key = %v, "%v" - "%v"`, i.Key, item.Comment, i.Comment)
				fmt.Fprintln(os.Stderr)
			}
			if len(item.Comment) == 0 {
				item.Comment = i.Comment
			}
			for k, v := range i.Localization {
				item.Localization[k] = v
			}
			dst[i.Key] = item
		} else {
			dst[i.Key] = i
		}
	}
}

func main() {
	find := flag.Bool("find", false, "find .strings files")
	prints := flag.Bool("print", false, "print .strings files")
	csv := flag.Bool("csv", false, "convert .strings to .csv")
	strings := flag.Bool("strings", false, "convert .csv to .strings")
	outputdir := flag.String("o", ".", ".strings output root directory")
	flag.Parse()

	if *find {
		f, err := FindStrings(".")
		if err != nil {
			log.Fatal(err)
		}

		for _, path := range f {
			fmt.Println(path)
		}
		return
	}

	if *prints {
		f, err := FindStrings(".")
		if err != nil {
			log.Fatal(err)
		}

		for _, path := range f {
			items, err := LoadStrings(path)
			if err != nil {
				log.Fatal(err)
			}

			for _, item := range items {
				fmt.Println(item)
			}
			fmt.Println()
		}
		return
	}

	if *csv {
		dst := map[string]Item{}

		f, err := FindStrings(".")
		if err != nil {
			log.Fatal(err)
		}

		for _, path := range f {
			items, err := LoadStrings(path)
			if err != nil {
				log.Fatal(err)
			}
			MergeItems(dst, items)
		}

		items := make([]Item, len(dst))
		i := 0
		for _, item := range dst {
			items[i] = item
			i++
		}

		var out io.Writer
		if len(flag.Args()) == 0 {
			out = os.Stdout
		} else {
			filename := flag.Arg(0)
			file, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			out = file
		}

		err = WriteCsv(out, items)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if *strings {
		var in io.Reader
		if len(flag.Args()) == 0 {
			in = os.Stdin
		} else {
			filename := flag.Arg(0)
			file, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			in = file
		}

		items, err := LoadCsv(in)
		if err != nil {
			log.Fatal(err)
		}

		err = WriteStrings(*outputdir, items)
		if err != nil {
			log.Fatal(err)
		}

		return
	}
}
