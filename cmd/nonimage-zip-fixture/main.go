// nonimage-zip-fixture 複製原版 ZIP，但排除所有已由 release PNG／JSON 取代的圖像 member。
// 這份輸出只供測試，證明 runtime 沒有暗中退回原版圖像；不得作為發行素材。
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var imageMember = regexp.MustCompile(`^(?:8X8D[1-6]|BIGPIC[126]|BODY[2-6]|CBODY|CHEAD|COMSPR|CPIC[1-6]|DUNGCOM|HEAD[2-6]|PIC[1-6]|RANDCOM|SKY|SPRIT[1-6]|TILES|TITLE|WALLDEF[2-6]|WILDCOM)\.DAX$`)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: nonimage-zip-fixture ORIGINAL.zip OUTPUT.zip")
	}
	input, err := zip.OpenReader(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(os.Args[2]), 0o755); err != nil {
		log.Fatal(err)
	}
	output, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	w := zip.NewWriter(output)
	kept, removed := 0, 0
	for _, source := range input.File {
		name := strings.ToUpper(filepath.Base(source.Name))
		if imageMember.MatchString(name) {
			removed++
			continue
		}
		header := source.FileHeader
		header.Method = zip.Deflate
		destination, err := w.CreateHeader(&header)
		if err != nil {
			log.Fatal(err)
		}
		h, err := source.Open()
		if err != nil {
			log.Fatal(err)
		}
		_, copyErr := io.Copy(destination, h)
		closeErr := h.Close()
		if copyErr != nil {
			log.Fatal(copyErr)
		}
		if closeErr != nil {
			log.Fatal(closeErr)
		}
		kept++
	}
	if err := w.Close(); err != nil {
		output.Close()
		log.Fatal(err)
	}
	if err := output.Close(); err != nil {
		log.Fatal(err)
	}
	if removed != 51 {
		log.Fatalf("removed image members=%d, want 51", removed)
	}
	fmt.Printf("kept=%d removed_images=%d output=%s\n", kept, removed, os.Args[2])
}
