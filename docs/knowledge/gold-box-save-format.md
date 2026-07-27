# Gold Box DOS 存檔與角色年齡欄位

這份文件集中記錄目前對 Curse of the Azure Bonds DOS save bundle 的可重用邊界。
它是 raw-preserving reverse-engineering reference，不是宣稱所有未知 bytes 都已解碼的
完整存檔編輯器。

## 檔案 bundle

- `SAVGAM?.DAT`：slot 的遊戲狀態與 party name references。
- `CHRDAT{slot}{1..6}.SAV`／`.GUY`：角色 record；實際載入時以作品的命名與 slot
  adapter 組合。
- 同 basename 的 `.FX`（effects）與 `.SWG`（items）是 optional sidecars。

修改既有角色時應保留整個 record，只 patch 已有證據的欄位；不要以 remake JSON 或
`SAVGAM` prefix 的長度推測 `.SAV/.GUY` 的完整 schema。

## 已確認的角色欄位

offset 以角色 record 起點計算，整數為 little-endian：

| offset | 欄位 | 說明 |
| --- | --- | --- |
| `0x00` | name length | 名稱長度 |
| `0x01..0x0F` | name | 固定名稱區 |
| `0x10..0x1B` | abilities | 六項能力值 |
| `0x74` | race | `0` monster、`1` dwarf、`2` elf、`3` gnome、`4` half-elf、`5` halfling、`6` half-orc、`7` human |
| `0x75` | class | `0` cleric、`1` druid、`2` fighter、`3` paladin、`4` ranger、`5` magic-user、`6` thief |
| `0x76..0x77` | age | signed little-endian 年齡 |
| `0x78` | max HP | 角色最大生命值 |
| `0x0DF..0x0E3` | saving throws | 五項 save 值 |
| `0x186` | saving bonus | 作品 record 的 saving bonus 欄位 |

因此，若只要修改年紀，目標是 `.SAV/.GUY` 的 `0x76` two-byte word；例如 25 歲的
little-endian bytes 是 `19 00`。修改前應先備份，並以 hex dump／parser round-trip 驗證
race、class、name、HP 與未知 bytes 沒有被改動。

## 年齡判讀

半獸人 raw race 是 `6`。目前 reference class-index 對應的起始年齡是：

- cleric：`20 + 1d4`
- fighter：`13 + 1d4`
- thief：`20 + 2d4`

這個生成表與 runtime age effects 是兩件事：`race_ages[race][class]` 只負責建立新角色，
既有 record 的 `0x76..0x77` 應直接讀寫。game clock 的 slot-6 年份進位則是另一條
runtime path；不要用 clock word 或 age bracket threshold 取代角色 record age。

## ECL shared flags

反組譯 reference 另外確認 ECL shared memory `$4C00` 起的 flag，映射到 `SAVGAM` 時使用
`((ecl_address - 0x4C00) * 2) + 0x201`。這個公式只適用已確認的 shared-flag window，
不能推廣成整個 save file 的通用 offset。

## 可重用修改流程

後續 Pool、Secret、Savage Frontier 等 Gold Box 作品可沿用以下 adapter contract；目前 remake
的 `ALTER → RENAME` 已使用同一條名稱 writeback 路徑：

1. 先辨識 slot 與 player sidecar，建立 backup。
2. 讀入 raw bytes，只修改已驗證 offset。
3. 以 sibling staging file 寫出，再替換目標檔；失敗時保留原檔。
4. 重新 parse 並檢查 age／race／class／name，未知 bytes 做 byte-for-byte preservation；角色
   ID 與 `.SAV/.GUY` sidecar basename 不因 rename 改變。

本 repo 已有 DOS player parser 與 SAVGAM staged writer；完整多職業欄位、所有 sidecar
schema、刪除／重排角色的原版副作用仍是未完成項目。

證據來源：本 repo 提供的 `Curse-of-the-Azure-Bonds_Manual_DOS_EN.pdf` 與 RefCard
說明操作流程，但沒有上述 raw offset 表；offset、race/class mapping 與 ECL table
operand 則依 [CoAB technical save-format guide](https://gamefaqs.gamespot.com/pc/564786/curse-of-the-azure-bonds/faqs/78365)
及 `/tmp/coab-reference` 的反編譯 reference 交叉核對。
