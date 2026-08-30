// Package gamecorpus 把「從原版 image 開一局」需要的資料集中在一處：六個 ECL
// 成員、六章的怪物表與物品表、六組 GEO、以及語系檔。
//
// ★ 存在的理由：盤點用的 `cmd/*` 各自抄一份載入流程，少載一樣東西的症狀都是
// **安靜地少一塊內容**——少載怪物表那一段就不開戰、少載物品表那一段在入口就
// 跑不動。集中一份，漏就是三支一起漏，看得見。
package gamecorpus

import (
	"archive/zip"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"os"
	"strings"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// Corpus 是開一局需要的全部原版資料。
type Corpus struct {
	Blocks  map[uint8][]byte
	Records map[uint8]map[uint8]monster.Record
	Items   map[uint16][]monster.ItemRecord
	Geo     map[uint8]geo.Catalog
	Catalog locale.Catalog
}

// Load 讀 image 與語系檔。六章的東西**全部**都載，不因為「這一段用不到」而跳過。
func Load(imagePath, localePath string) (Corpus, error) {
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		return Corpus{}, err
	}
	defer archive.Close()

	data := Corpus{
		Blocks:  map[uint8][]byte{},
		Records: map[uint8]map[uint8]monster.Record{},
		Geo:     map[uint8]geo.Catalog{},
	}
	itemPayloads := map[uint8][]byte{}
	for chapter := 1; chapter <= 6; chapter++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", chapter))
		if payload == nil {
			return Corpus{}, tooltext.Errorf("h.f74a43cb81d6", chapter)
		}
		parsed, err := dax.Parse(payload)
		if err != nil {
			return Corpus{}, err
		}
		for _, block := range parsed {
			data.Blocks[block.Entry.ID] = block.Data
		}
		if monsters := memberPayload(archive, fmt.Sprintf("MON%dCHA.DAX", chapter)); monsters != nil {
			records, err := parseMonsters(chapter, monsters)
			if err != nil {
				return Corpus{}, err
			}
			data.Records[uint8(chapter)] = records
		}
		if items := memberPayload(archive, fmt.Sprintf("ITEM%d.DAX", chapter)); items != nil {
			itemPayloads[uint8(chapter)] = items
		}
		if maps := memberPayload(archive, fmt.Sprintf("GEO%d.DAX", chapter)); maps != nil {
			catalog := geo.NewCatalog()
			if err := catalog.AddDAX(uint8(chapter), maps); err != nil {
				return Corpus{}, fmt.Errorf("GEO%d: %w", chapter, err)
			}
			data.Geo[uint8(chapter)] = catalog
		}
	}
	data.Items, err = game.ParseTreasureItemBlocks(itemPayloads)
	if err != nil {
		return Corpus{}, err
	}
	localeData, err := os.ReadFile(localePath)
	if err != nil {
		return Corpus{}, err
	}
	data.Catalog, err = locale.Load(localeData)
	if err != nil {
		return Corpus{}, err
	}
	return data, nil
}

// NewParty 開一局並建一名角色，停在建角完成的狀態。
func (c Corpus) NewParty() (game.State, error) {
	state := game.NewStateFromECLBlocks(c.Catalog, c.Blocks, 0x50)
	for chapter, records := range c.Records {
		state.SetMonsterRecordsForECL(chapter, records)
	}
	state.SetTreasureItemBlocks(c.Items)
	if err := state.OpenCharacterCreation(); err != nil {
		return state, err
	}
	if err := state.AddCreationCharacter(0); err != nil {
		return state, err
	}
	if err := state.FinishCharacterCreation(); err != nil {
		return state, err
	}
	return state, nil
}

// BoostParty 把隊伍撐到足以走完內容盤點的程度。**只給盤點用**：段測試不准這樣
// 做，那會把「這一段打得贏嗎」偷偷換成「這一段演得出來嗎」。
func BoostParty(state *game.State) error {
	party := state.PartyFighters()
	if len(party) == 0 {
		return tooltext.Errorf("h.b54a6a6a8b20")
	}
	for index := range party {
		party[index].HitPoints, party[index].MaxHitPoints = 999, 999
		party[index].ArmorClass = -10
		party[index].AttackBonus = 100
		party[index].DamageDiceCount, party[index].DamageDiceSides = 1, 1
		party[index].DamageBonus = 100
		party[index].AttacksPerTurn = 8
		party[index].InitiativeBonus = 100
	}
	return state.SetParty(party)
}

func parseMonsters(chapter int, payload []byte) (map[uint8]monster.Record, error) {
	blocks, err := dax.Parse(payload)
	if err != nil {
		return nil, err
	}
	records := map[uint8]monster.Record{}
	for _, block := range blocks {
		record, err := monster.Parse(block.Data)
		if err != nil {
			return nil, fmt.Errorf("MON%dCHA block %#02x: %w", chapter, block.Entry.ID, err)
		}
		records[block.Entry.ID] = record
	}
	return records, nil
}

func memberPayload(archive *zip.ReadCloser, member string) []byte {
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, member) {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil
		}
		defer reader.Close()
		payload, err := io.ReadAll(reader)
		if err != nil {
			return nil
		}
		return payload
	}
	return nil
}
