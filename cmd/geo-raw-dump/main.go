// Command geo-raw-dump 把一張 GEO 圖的**四個平面原樣**印出來。
//
// ★ 存在的理由：`cmd/geo-wall-grid` 印的是解讀過的牆面指紋，答不了「這一格
// 到底有沒有資料」。要分辨「地圖本來就長這樣」與「解析漏了一塊」，得看
// 未經解讀的位元組。
//
// 版面（`internal/geo`／engine `geometry.Parse`）：兩個 byte 的長度前綴
// （`00 04` ＝ 0x400）之後是四個 0x100 的平面——
//
//	平面 0   每格一個 byte，高 nibble ＝ 北面牆型、低 nibble ＝ 東面
//	平面 1   高 nibble ＝ 南面、低 nibble ＝ 西面
//	平面 2   地形
//	平面 3   細節（四個 2-bit 欄位）
//
// ⚠ **四個平面不是四個方向**：前兩個各自打包兩個方向。把它讀成「北東南西各
// 一個平面」會得到自洽但錯的牆面圖。
//
// ⚠ **地形 0 不是「不能站」**：全作 4,096 格裡有 1,042 格是 0，而
// `geo6-b42` 有 204／256 格是 0 卻逐格全等（spec 1185）。
//
// 用法：
//
//	go run ./cmd/geo-raw-dump -set 5 -block 51
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	set := flag.Int("set", 5, tooltext.Text("h.7aaf8dbad92f"))
	block := flag.Int("block", 0x33, tooltext.Text("h.e3d3f40d66aa"))
	flag.Parse()
	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	name := fmt.Sprintf("GEO%d.DAX", *set)
	var payload []byte
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		handle, openErr := file.Open()
		if openErr != nil {
			log.Fatal(openErr)
		}
		payload, _ = io.ReadAll(handle)
		handle.Close()
	}
	blocks, err := dax.Parse(payload)
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range blocks {
		if int(b.Entry.ID) != *block {
			continue
		}
		data := b.Data
		fmt.Print(tooltext.Format("h.2d534566a5fe", b.Entry.ID, len(data)))
		body := data
		if len(data) == 0x402 {
			fmt.Print(tooltext.Format("h.a22d084f5c7d", data[0], data[1]))
			body = data[2:]
		}
		names := []string{tooltext.Text("h.fcfd1061f25a"), tooltext.Text("h.8c461f636887"), tooltext.Text("h.637a28261eea"), tooltext.Text("h.c1da405769bd")}
		for plane := 0; plane < 4; plane++ {
			fmt.Printf("== %s ==\n", names[plane])
			for y := 0; y < 16; y++ {
				row := make([]string, 0, 16)
				for x := 0; x < 16; x++ {
					row = append(row, fmt.Sprintf("%02X", body[plane*0x100+y*16+x]))
				}
				fmt.Printf("y=%2d %s\n", y, strings.Join(row, " "))
			}
		}
	}
}
