# 1194 — 原作的全域熱鍵：完整的四個，remake 接了兩個

狀態：`READY`（分母已封閉：熱鍵處理是一支常式，四個分支全部認出來了）

## 為什麼這件事有分母

原作的全域熱鍵**不是散在各個畫面**，而是集中在一支常式：PC-98 `sub_18036`
（linear `18036h`）。它讀一個鍵，然後依序比四個值；其餘的原樣回傳給呼叫端。
所以「原作有幾個全域熱鍵」是個**可以問完**的問題——四個，沒有第五個。

```
al := 讀鍵()
cmp al, 13h  → Ctrl+S  音效開關
cmp al, 0Fh  → Ctrl+O  音樂開關
cmp al, 02h  → Ctrl+B  BIGAIM（範圍瞄準游標）
cmp al, 16h  → Ctrl+V  VISUALTYPE（PC-98 螢幕模式）
回傳 al
```

⚠ 內部碼就是 **ASCII 控制碼**：`13h` ＝ Ctrl+S，和 DOS 參考卡上寫的
`CTRL S : Toggles sound on and off (may be used at any time).` 對得起來。
這個對照是後面三個推 Ctrl+O／Ctrl+B／Ctrl+V 的依據。

★ 同一支常式也是**設定畫面**（`sub_1259F`）用的：那個畫面把四個設定各印一行
ON／OFF，按鍵一樣走 `sub_18036`。所以熱鍵與設定畫面是同一組東西的兩個入口。

## 四個設定

| 熱鍵 | 變數 | 是什麼 | 誰真的讀它 | remake |
|---|---|---|---|---|
| Ctrl+S | `SOUNDTYPE` `8B58h` | 音效開關（`2` ＝ 靜音）| `SOUNDFX` 開頭 | ✅ `ToggleSoundSwitch` |
| Ctrl+O | `MUSICSW` `8BE3h` | 音樂開關（`1` ＝ 關）| 派曲常式 `sub_18AA7` 第一關 | ✅ `ToggleMusicSwitch` |
| Ctrl+B | `BIGAIMON` `8BEEh` | 範圍瞄準游標 | COMSTUFF ×2、TACMAP ×2 | — 未做 |
| Ctrl+V | `VISUALTYPE` `8BF2h` | PC-98 螢幕模式 | INIT（overlay-11）×1 | — 不適用 |

音訊那兩個的細節在 spec 1192。

### Ctrl+B：範圍瞄準游標

`BIGAIMON` 有四個讀取點，都在戰鬥側：

```
COMSTUFF 3307h   mov al,[BIGAIMON]; xor al,[BIGAIMOLD]; or al,al; jz 略過
                 ⇒ 「設定變了」才重畫：變了用屬性 0FFh，沒變用 3
                 然後 BIGAIMOLD := BIGAIMON
TACMAP   049Bh   cmp [BIGAIMON],0; jz 跳過 125h 位元組
TACMAP   0E2Ah   cmp [BIGAIMON],0; jz 跳過 3B5h 位元組
```

兩處 TACMAP 的守衛後面接的都是 `SPECIALCASTNUM`／`SPECIALAIMTYPE`（`BE2Dh`／
`BE2Ch`）的分支，所以這個設定管的是**施放範圍法術時要不要把作用範圍畫出來**，
不是一般的游標大小。

⚠ **這不是「還沒接的音訊動作」**，是一項還沒做的戰鬥介面功能，要等 TACMAP 的
範圍瞄準本身做到那個程度才談得上。列在這裡是為了讓分母是完整的四個。

### Ctrl+V：螢幕模式

`sub_18194` 是純 PC-98 硬體：`out 6Ah`／`out 68h`／`int 18h`（PC-98 的 CRT
BIOS，**不是** IDA 註解寫的 ROM BASIC），依設定送 `304h` 或 `334h` 兩組 GDC
參數。這是**顯示器時序**，跨平台 remake 沒有對應物——不做不是缺漏。

## 實作要求（兩個已接的）

- **擺在所有模式前面。** 參考卡的 `may be used at any time` 是實作要求，
  連地圖預覽那種提早 `return` 的畫面也要管得到。
- **按鍵要被吃掉。** 原作是在讀鍵的地方攔下來，不會再往下傳；少了這一步，
  同一幀的 `S` 會被戰鬥選單當成別的指令。
- `BIGAIMCHG`（`8BF0h`）在這一版**寫了但沒有人讀**；真正拿來比對的是
  `BIGAIMOLD`（`8BEFh`）。照著名字接會接錯邊。

## 這個數字證明不了什麼

- **PC-98 版面。** DOS 版沒有除錯符號，四個裡只有 Ctrl+S 有參考卡佐證；
  其餘三個是 PC-98 這一支常式讀出來的，不保證 DOS 版有同樣的鍵。
- 「認出來了」不表示行為完全相同——Ctrl+B 只推到「管範圍瞄準的顯示」，
  沒有逐像素比對過。

## 相關

- spec 1192（音訊生命週期）、1185（第一人稱畫面）
- `cmd/azure-bonds-game/keys.go`：`globalAudioKeys`
- `TestOriginalGlobalHotkeysAreAccountedFor`
