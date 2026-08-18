# 1133 — 第一份原版存檔樣本；`CHRDAT` 檔名欄是 Pascal 短字串

- 證據等級：`exact`（樣本由原版 DOS 版自己寫出；檔名欄的形狀在樣本的位元組裡
  直接看得到，並與 spec 1072 從 PC-98 逐條讀到的 `string[40]` 一致）
- 上游 spec 102（「那一層必須先取得實際 sample bytes」）、spec 1072（存檔的
  16 個區塊）、spec 181（DOS 側的固定前綴）、spec 1121（原版與 remake 兩條匯入路）

## 一、拿到原版自己寫的存檔樣本

在 Docker/DOSBox 裡用遊戲自己的 `START.EXE STING Wooden` 測試模式（spec 530）
建一個角色再存檔，原版寫出兩個檔案，收在
[`docs/reference/original-dos/save-samples/`](../reference/original-dos/save-samples/)：

| 檔案 | 長度 | 對得上什麼 |
|---|---:|---|
| `BOB.GUY` | 422 | `DOSPlayerRecordSize`（spec 1115 的 422 bytes 角色記錄）|
| `BOB.FX` | 27 | `3 × 9`，`AffectRecordSize` 的整數倍 |

spec 102 當年寫「那一層必須先取得實際 sample bytes 與 header/record boundary
證據」。這一組就是——而且是**原版產生**的，不是 remake 自己寫的：兩者版面相同，
拿 remake 的輸出當樣本等於沒測（spec 1121 同一個陷阱）。

`TestOriginalDOSPlayerSampleDecodes` 用 `ParseOriginalDOSPlayerFiles` 讀這一組，
釘住名字、最大 HP 與力量都解得出來。

## 二、★ `CHRDAT` 檔名欄是 Turbo Pascal 的 `string[40]`

樣本第一眼就看得到：`BOB.GUY` 的開頭是 `03 'B' 'O' 'B'`——**長度位元組在前**。
spec 1072 在 PC-98 側量到的存檔版面也是同一個形狀：第 16 塊 `148h` bytes ÷
`29h` ＝ 8 筆，每筆 `1 + 40`。

remake 的 `EncodeSAVGAM` 只負責把 41 bytes 原樣搬進去，而兩個呼叫端
（`SaveSAVGAMSlot`、`savgamContainerForSave`）都是直接丟名字的原始位元組進去，
**沒有長度位元組**。原版讀到那一欄會把第一個字元當成長度去組檔名。

⚠ **remake 自己讀自己的槽察覺不到**：`loadSAVGAMSlot` 是照 `CHRDAT{槽}{n}`
算出檔名，根本沒讀這一欄。所以這種錯只有**拿原版當 oracle** 才驗得出來——
自洽的往返測試對它是盲的。

修法是加一支 `save.SAVGAMCharacterRef(name)` 專門包這 41 bytes，兩個呼叫端改用它。

## 三、⚠ 前綴長度：兩份規格差 `41h`

| 規格 | 平台 | ECL 區塊 | 檔案總長 |
|---|---|---:|---:|
| spec 181 | DOS | `1E00h` ＝ 7680 | 13,149 |
| spec 1072 | PC-98 | `1E41h` ＝ 7745 | 13,214 |

差的 65 bytes 全落在那一塊。兩者的證據等級不同：spec 1072 是**逐條讀完、
無匯出破洞**；spec 181 的來源是 `engine/ovr017.cs` 的 **decompiler 輸出**，
而本專案的規則明寫「IDA／decompiler 的輸出本身不等於證明」（`AGENTS.md` §3）。

本輪**沒有**改 `SAVGAMECLMemorySize`：要改得先在 DOS 側逐條讀 `SaveGame`，
或拿到一份原版寫出來的 `savgam?.dat` 量長度。`cmd/dos-save-export` 因此留了
`-pad-ecl` 讓兩種長度都試得出來。

## 四、還沒成功的：讓原版讀進合成的存檔

`cmd/dos-save-export` 會產生 `savgam<槽>.dat` ＋ 角色檔。目的是取 oracle：
讓原版站在指定的格子上，再拍第一人稱畫面來比對牆磚選圖（spec 1131 的收尾）。

目前原版**拒絕**這份檔案——按下槽的字母之後直接回主選單。已經排除的變因：

- 前綴長度（13,149 與 13,214 兩種都試過）
- 角色檔（改用原版自己寫的 `A.GUY`，而不是 remake 產生的記錄）
- 檔名欄的長度位元組（本輪修正之後仍然不讀）

還沒查的：`GameState`／`LastGameState`／三組 `blockId`／`setId` 的合法值、
Area1／Area2 裡是否有原版必須看到的欄位、以及原版開角色檔時用的副檔名
（獨立角色是 `.GUY`，存檔裡的那八筆用哪一個還沒證）。

## 明確不宣稱

- 沒有宣稱 DOS 側的 ECL 區塊是 `1E00h` 或 `1E41h`；本規格只記錄兩份證據的落差
  與各自的證據等級。
- 沒有宣稱 `BOB.GUY` 的 422 bytes 每一格都解對——只驗了名字、最大 HP 與力量
  三個欄位讀得出來且合理。
- 沒有宣稱 `.FX` 的 27 bytes 是三筆有意義的效果；只驗了長度是 9 的倍數且解析不報錯。
- 沒有宣稱原版拒絕合成存檔的原因；上面列的是**已排除**的變因，不是結論。
