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

## 三、前綴長度是平台差異

| 平台 | ECL 區塊 | 檔案總長 | 來源 |
|---|---:|---:|---|
| DOS | `1E00h` ＝ 7680 | 13,149 | 原版自己寫出來的 `savgam?.dat` 量到；overlay-16 `sub_3748` 的 `39BC b8001e mov ax, 1E00h` |
| PC-98 | `1E41h` ＝ 7745 | 13,214 | spec 1072 逐條讀完 |

差的 65 bytes 全落在那一塊，兩份規格各自對自己的平台正確。
`SAVGAMECLMemorySize` 維持 `1E00h`（本專案的執行檔是 DOS 版）。

## 四、原版讀不讀得進合成的存檔

讀得進去。合成檔與原版存檔逐區塊比對只差兩處（`Area2 + 67Ch` 會被載入端
清成 0，以及角色檔名那一欄），版面本身沒有問題。

⚠ **檔名欄指到不存在的檔案時，原版不報錯，直接跳過那名角色**（overlay-16 `3B84`）。
八筆全跳過就得到空隊伍，而空隊伍會退回「沒有隊伍」的主選單——畫面上看起來
就是「按下槽的字母之後回主選單」，與拒絕讀檔一模一樣。**沒有錯誤訊息的失敗，
症狀會長得像另一個問題。**

副檔名是 `.sav`（連帶要有同名的 `.FX`），不是 `.GUY`——原版存檔時會把
`<名字>.GUY`／`.FX` 改名成 `CHRDAT<槽><序>.SAV`／`.FX`。

細節與取 oracle 的完整流程見
[spec 1134](1134-original-first-person-oracle.md)。

## 明確不宣稱

- 沒有宣稱 `BOB.GUY` 的 422 bytes 每一格都解對——只驗了名字、最大 HP 與力量
  三個欄位讀得出來且合理。
- 沒有宣稱 `.FX` 的 27 bytes 是三筆有意義的效果；只驗了長度是 9 的倍數且解析不報錯。
- 沒有宣稱 `CHRDAT<槽><序>` 之外的檔名慣例；驗過的是單人隊伍存到槽 A／B。
