# Gold Box PARLAY 五態度分支契約

本文件只整理《Curse of the Azure Bonds》ECL5 block `0x33` 龍群交涉分支，
並把其中可跨 Gold Box 作品重用的 VM 契約與作品專屬資料分開。offset 均為
block 解壓後、略過兩-byte prefix 的 ECL payload 相對 offset；`0x8000 + offset`
才是 script 內 word address。

## CoAB block `0x33` 龍群分支

外層四項 VERTICAL MENU 的 `PARLAY WITH THE DRAGONS` 分支抵達
`+0x05E7`；`+0x05E7 PICTURE 0x36` 後，`+0x05EA` 是真正的
`PARLAY (0x2C)` opcode。

原始 ECL5.DAX 在 `+0x05EA` 的六個 operands 是：

```text
PARLAY 1, 0, 0, 0, 1, [0x7F79]
```

前五值依既有 Gold Box PARLAY 指令反組契約，按零起算 index 對應以下原始
label。DAX command 只保存五個結果值，不重複保存英文 label；因此 label
順序是 opcode／menu 契約，`[1,0,0,0,1]` 才是本 block 的原始作品資料。

| index | 原始 label | 寫入 `0x7F79` | script 判斷與 effective target | 結果 |
|---:|---|---:|---|---|
| 0 | `HAUGHTY` | 1 | `+0x05FB` 比較為 1，`+0x0602 GOTO 0x84B9` → `+0x04B9` | 14 黑龍戰 |
| 1 | `SLY` | 0 | 比較不成立，fall-through → `+0x0606`；對話後 `+0x0672 GOTO 0x8564` → `+0x0564` | 說服龍群，接德拉坎德羅斯守軍 |
| 2 | `MEEK` | 0 | fall-through → `+0x0606`；再由 `+0x0672` → `+0x0564` | 說服龍群，接德拉坎德羅斯守軍 |
| 3 | `NICE` | 0 | fall-through → `+0x0606`；再由 `+0x0672` → `+0x0564` | 說服龍群，接德拉坎德羅斯守軍 |
| 4 | `ABUSIVE` | 1 | `+0x05FB` 比較為 1，`+0x0602` → `+0x04B9` | 14 黑龍戰 |

`+0x05F8 PICTURE 0xFF` 會在判斷前關閉交涉圖。`+0x05FB` 比較
`[0x7F79] == 1`，所以這裡不是五條各自帶 address 的 jump table，而是五個
態度先映射成二值，再由共同 branch 分流。

### target 證據

- `+0x04B9 SETUP MONSTER 0x35,1,0x35`；
- `+0x04F4 CLEARMONSTERS`；
- `+0x04F5 LOAD MONSTER 0x35,14,0x35`；
- `+0x04FC COMBAT`。

以上直接證明 `HAUGHTY`、`ABUSIVE` 導向 14 名 MON5 `0x35`
`BLACK DRAGON`，而不是僅由對話文字推測。

說服路徑則由：

- `+0x0606 PICTURE 0x35`；
- `+0x0609..+0x0662` 龍群表示已被說服，將離開並讓隊伍與
  Dracandros 解決爭端；
- `+0x0672 GOTO 0x8564`，即 payload `+0x0564`。

`+0x0564` 是共用的 Dracandros 守軍 continuation：選圖後播放
「TROOPS DEFEND ME!」與 Dracandros 逃下樓的文字，再於
`+0x05CC..+0x05E2` 建立並進入伊弗利特一名、黑暗精靈戰士兩名、
黑暗精靈法師一名的戰鬥。因此 `SLY`、`MEEK`、`NICE` 都導向同一守軍
encounter，沒有三套不同對話 target。

## 可跨 Gold Box 重用的 opcode／continuation 契約

`PARLAY (0x2C)` 是 VM input boundary，不是顯示完選單便結束事件的
frontend action。可重用 runner 應遵守：

1. 解出六個 operands：index `0..4` 的五個 result expression，以及第六個
   destination memory word。
2. 固定保留原始態度順序 `HAUGHTY / SLY / MEEK / NICE / ABUSIVE`；
   在地化只能替換顯示文字，不得改變 index。
3. 尚無 selection 時，在此 opcode 暫停，保存目前 block、PC、memory 與
   selection cursor；不得由 State／UI 先生成作品外的泛用交涉結果。
4. resume 時只消耗一個 selection。以所選 index 求值對應 operand，將結果
   寫入第六 operand 指向的 memory，再從 opcode 的 `Next` 繼續執行。本例
   `Next` 是 `+0x05F8`，destination 是 `0x7F79`。
5. branch target 屬於後續 ECL script，不屬於 PARLAY opcode。runner 不可把
   result `0/1` 當成 code address，也不可在共用引擎硬編碼「某態度必定戰鬥」；
   本例的結果語意由 `+0x05FB..+0x0602` 決定。
6. `PARLAY` opcode 與文字恰好含 `PARLAY` 的普通 HORIZONTAL／VERTICAL MENU
   必須分辨。普通 menu 的 label 沒有權力觸發此五態度契約。

這個分層允許其他 Gold Box 遊戲沿用相同 pause／resume、index、result-write
機制，同時由各作品自己的五個 operands 與後續 ECL control flow 決定結果。

## 尚未知／不可宣稱

- 這個 block 只證明 `[1,0,0,0,1]`；不可宣稱所有 Gold Box 遊戲或所有遭遇
  都使用相同態度結果。
- DAX 未在此 command 內保存五個英文 label；不可把 label 順序誤稱為
  block-local 字串證據。
- 尚未由本分支證明談判者如何選擇、Charisma／reaction roll、speaker、
  monster intelligence／morale、隨機修正或態度數值的全域公式；不可替
  `0/1` 擴張出這些語意。
- 本分支三個成功態度匯流，不代表 `SLY`、`MEEK`、`NICE` 在其他 encounter
  等價；兩個戰鬥態度亦同。
- `+0x04FD` 之後的黑龍重戰 gate、龍心與酸液流程已由
  [規格 318](../spec/318-wizard-tower-black-dragons-heart.md) 另行核對；
  它們是共用 hostile target 的後續，不應反推成 PARLAY opcode 的全域語意。
