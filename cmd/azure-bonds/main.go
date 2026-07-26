package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "original DOS image ZIP")
	member := flag.String("member", "ECL1.DAX", "DAX member to inspect")
	flag.Parse()
	data, err := zipMember(*image, *member)
	if err != nil {
		log.Fatal(err)
	}
	blocks, err := dax.Parse(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %d blocks\n", *member, len(blocks))
	for _, block := range blocks {
		fmt.Printf("block %d: %d decoded bytes\n", block.Entry.ID, len(block.Data))
	}
}

func zipMember(path, member string) ([]byte, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name != member {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	}
	return nil, fmt.Errorf("member %q not found", member)
}
