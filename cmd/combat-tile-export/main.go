// combat-tile-export 把戰鬥地形圖集（DUNGCOM／WILDCOM／RANDCOM）的每一格
// 匯出成 PNG。
//
// ★ 存在的理由是「用資料回答，不要用印象回答」。第一次用它是為了驗
// game pack 宣告的雲霧格：RANDCOM 2 與 4 到底是不是雲（是——2 是藍白、
// 4 是綠色，spec 1128）。⚠ 雲霧格在 RANDCOM，不在 COMSPR；拿 COMSPR 的
// 同號區塊去看會看到箭矢與投射物，然後得到「宣告錯了」這個錯結論。
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
)

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	out := flag.String("out", "", "output directory")
	flag.Parse()
	if *out == "" {
		log.Fatal("-out is required")
	}
	reader, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	for _, source := range []string{"DUNGCOM", "WILDCOM", "RANDCOM"} {
		var data []byte
		for _, file := range reader.File {
			if !strings.EqualFold(file.Name, source+".DAX") {
				continue
			}
			handle, err := file.Open()
			if err != nil {
				log.Fatal(err)
			}
			data, err = io.ReadAll(handle)
			handle.Close()
			if err != nil {
				log.Fatal(err)
			}
		}
		blocks, err := dax.Parse(data)
		if err != nil || len(blocks) != 1 {
			log.Fatalf("parse %s.DAX: blocks=%d err=%v", source, len(blocks), err)
		}
		set, err := gfx.ParseCombatTiles(blocks[0].Data)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s tiles=%d\n", source, len(set.Tiles))
		for index, tile := range set.Tiles {
			rgba, err := tile.RGBA(0, gfx.EGA16)
			if err != nil {
				log.Fatal(err)
			}
			path := filepath.Join(*out, fmt.Sprintf("%s-%02d.png", source, index))
			handle, err := os.Create(path)
			if err != nil {
				log.Fatal(err)
			}
			if err := png.Encode(handle, rgba); err != nil {
				log.Fatal(err)
			}
			handle.Close()
		}
	}
}
