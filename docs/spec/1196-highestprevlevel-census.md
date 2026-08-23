# `+0E6h` `HIGHESTPREVLEVEL`：全部 15 處讀取的普查

狀態：`READY`

## 結論

角色記錄 `+0E6h` 是 **`HIGHESTPREVLEVEL`**（PC-98 Borland 符號，型別
`LEVELTYPE`，1 byte），夾在 `+0E5h HIGHESTLEVEL` 與 `+0E7h LOSTLEVELS` 之間
（`docs/audit/pc98-record-layouts.md`）。它是**雙職角色的舊職等級門檻**，
不是「多職角色的現行等級」。

## 兩平台的讀取普查

掃 `cmp al, es:[di+0E6h]`（`26 3A 85 E6 00`）與 `mov al, es:[di+0E6h]`
（`26 8A 85 E6 00`）：

| 模組 | `cmp` | `mov` |
|---|---:|---:|
| `TEMPLE` | 1 | 0 |
| `COMPTACT` | 1 | 0 |
| `GEN` | 1 | 0 |
| `SPELLS` | 8 | 0 |
| `EFFECTS` | 2 | 1 |
| `TRAINING` | 1 | 0 |
| **合計** | **14** | **1** |

⚠ **DOS 與 PC-98 逐模組、逐數量完全相同**（位移不同、計數一致）。
兩份獨立的建置給出同一張表，這比任何單一掃描的自證都有力。

⚠ 這是位元組線性掃描，`es:[di+0E6h]` 的 `di` 理論上可以指向別的結構 ⇒ 有偽陽性
的可能。但六個模組全部是角色相關的（神殿、戰術、通用、法術、效果、訓練），
而且兩平台一致 ⇒ 當成同一個欄位是站得住的。

⚠ **不要用「少一個模式」的掃描下「只有兩處」的結論。** 第一版只掃了
`mov`＋幾個沒有 `es:` 前綴的形式，得到「PC-98 只有兩處」，而正確答案是 15 處。
`es:` 段前綴是 Borland 存取遠指標欄位的**常態**，漏掉它會系統性少報
（`cmd/dseg-writers` 的檔頭也記著同一個盲點）。

## 語意的錨點

`EFFECTS` 的 `overlay-23:1D1Dh`（PC-98）就是 spec 1079 那條累加 HP 的門檻，
逐指令對得上：

```
1D17  al := 等級（先被職業等級上限表 [6F4Dh] 夾過）
1D1D  cmp al, es:[di+0E6h]
1D22  jbe  → 不累加
1D27  al := es:[di+0E6h]
1D2C  等級 := 等級 − al
1D37  call 累加HP(職業, 等級差)
```

也就是 AD&D 的雙職規則：**只有超過舊職等級的那幾級才給 HP**。
spec 1092 那條「`角色^[0E6h] = 0` 才能編輯」是同一個語意的另一面：
沒有雙職過的角色才動得了。

## 兩件要記的事

- **spec 185 從頭到尾沒有提過 `+0E6h`。** `internal/party/dos_record_fields.go`
  先前把「多職角色的現行等級」這個讀法掛在 spec 185 名下，那個出處不存在。
  引用一個規格之前先在那份規格裡 grep 一次。
- **remake 曾經把 `data[0xE6]` 當成「顯示等級」**
  （`internal/party/dos_spell_record.go`，職業碼 8..16 時），而**原版的 15 處
  沒有一處是拿它去顯示的**，全部是門檻比較。第 751 輪改掉了，見下一節。

## 原版顯示的等級是 `PREVIOUSLEVEL[槽] + CURRENTLEVEL[槽]`（第 751 輪）

不需要實機量：**顯示常式自己講得很清楚**。`LIBRARY`（`overlay-19`）逐槽跑
0..7 組角色卡上的等級字串：

```
0374  cmp  byte es:[di+109h], 0     ; CURRENTLEVEL[槽] > 0 → 直接印
037A  jg   顯示
0384  call far …                    ; dl := 由角色算出來的一個門檻
0395  mov  al, es:[di+111h]         ; PREVIOUSLEVEL[槽]
039A  cmp  al, dl / jl 繼續判        ; 舊職等級 >= 門檻 → 這一槽不印
03AB  cmp  byte es:[di+111h], 0
03B1  jg   顯示 / 否則跳過
顯示:
0408  mov  al, es:[di+111h]         ; PREVIOUSLEVEL[槽]
040D  cbw / mov dx, ax
041A  mov  al, es:[di+109h]         ; CURRENTLEVEL[槽]
0420  add  ax, dx                   ; ★ 印出來的就是這個和
0422  push ax / call 數字轉字串
```

⇒ **一個槽印一列，數字是 `PREVIOUSLEVEL[槽] + CURRENTLEVEL[槽]`。**
`+0E6h` 在整支裡一次都沒出現——與上面那張普查表一致（`LIBRARY` 不在名單裡）。

掃描面的佐證（位元組直掃 `26 8A 85 <ofs>`）：

| 欄位 | 讀到的模組 |
|---|---|
| `+109h` CURRENTLEVEL | overlay-02、04、13、**19**、17、22、23、24、25 |
| `+111h` PREVIOUSLEVEL | overlay-02、09、13、**19**、17、22、23、24、25 |
| `+0E6h` HIGHESTPREVLEVEL | **只有 overlay-23（EFFECTS）** |

remake 這一側因此改成取各槽 `CURRENTLEVEL + PREVIOUSLEVEL` 的最大值當投影
（`DOSPlayerRecord.Level` 是單一純量，原作沒有這種東西）。純多職角色的
`PREVIOUSLEVEL` 是 0 ⇒ 結果與舊的 `max(職業等級)` 後備相同；**雙職角色才看得出
差別**，而那正是舊寫法唯一會錯的地方。`TestDualClassLevelComesFromPreviousPlusCurrent`
把三種情況（純多職、雙職、`+0E6h` 有值）都釘住。
