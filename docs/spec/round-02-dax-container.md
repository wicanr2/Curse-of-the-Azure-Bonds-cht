# 第二輪：DAX 容器與 ECL 文字取樣

狀態：`DRAFT`

## 證據來源

- 原始 `curseoftheazurebonds.zip` 的位元組樣本。
- Gold Box Explorer 的公開 DAX reader：其資料層採用 2-byte header offset、9-byte block entry 與 RLE 解碼；本專案以獨立 Python 工具重新實作並對本地樣本驗證。
- 社群研究指出不同 Gold Box 遊戲的 ECL 結構會有變體，因此本規格只針對 Curse of the Azure Bonds 的樣本下結論。

參考：[Gold Box Explorer DAX reader](https://github.com/bsimser/Gold-Box-Explorer/blob/master/src/Common/Plugins/Dax/DaxFile.cs)、[Gold Box 遊戲版本與 ECL 變體討論](https://forums.goldbox.games/index.php?topic=3065.0)。

## DAX 容器格式（目前可實作）

所有整數為 little-endian。

```text
u16 header_size_minus_2
HeaderEntry[(header_size_minus_2 / 9)]
byte data_area[]

HeaderEntry =
    u8  block_id
    u32 offset_from_data_area
    u16 decoded_size
    u16 packed_size
```

資料區起點為 `header_size_minus_2 + 2`。每個 block 的 packed bytes 位於 `data_start + offset`。

### RLE

- control byte 為 signed int8。
- `0..127`：後續 `control + 1` bytes 原樣複製。
- `-1..-128`：讀下一個 byte，重複 `-control` 次。
- 解碼必須剛好產生 `decoded_size` bytes；截斷或溢位視為格式錯誤。

## ECL 特化觀察

- `ECL1.DAX` 解析出 3 blocks（IDs `0x50`–`0x52`）。
- `ECL2.DAX` 解析出 4 blocks（IDs `1`–`4`）。
- block ID 是事件／區域資料的識別值，不等於 `ECL` 檔案序號。
- 部分解碼資料在 `0x80 length payload` 位置包含 6-bit packed text；目前工具可取樣，但尚未證明所有 `0x80` 都是字串。

## 尚未 READY 的原因

- ECL block 的事件記錄、opcode／資料欄位仍未分離。
- DAX block 的 RLE 已可驗證，但圖片、地圖與 ECL 的 payload 是不同上層格式。
- 文字抽取存在 false positive，需用實際畫面／反組譯交叉驗證。

## 驗收命令

```sh
python3 scripts/dax_dump.py curseoftheazurebonds.zip --member ECL1.DAX
python3 scripts/dax_dump.py curseoftheazurebonds.zip --member GEO2.DAX
python3 -m unittest tests/test_dax_dump.py
```
