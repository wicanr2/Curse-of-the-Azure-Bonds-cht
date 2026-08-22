# 1183 — 誰寫地圖暫存器：`720Fh`／`7210h`／`7211h` 的全域普查

- 證據等級：`exact`（DOS 執行檔逐位元組掃描，兩種掃描各自有正對照）
- 上游：spec 1150（`STOREVALUE` 的鏡射）、spec 1172（`ViewMirror` 與它的補丁）、
  spec 1163（SAVGAM 的區塊版面）
- 產物：`cmd/dseg-writers`、`docs/audit/dseg-writers-map-registers.md`

## 為什麼要做這件事

spec 1172 把 `ViewMirror.Block` 寫成「remake 這一側的補丁」，理由是：

> 換 block 的進場放置在原作是引擎做的、會覆蓋三格。

整個補丁的正當性靠這句話撐著，而這句話**從來沒有被驗證過**——它是從
「remake 拿掉那個比對之後入口位置會跑掉」反推的。所以先去問二進位。

## 結論

`720Fh`／`7210h`／`7211h` 在 DOS 版的**寫入者只有五處**：

| overlay | 單元 | 位址 | 什麼 |
|---|---|---|---|
| 07 | ECL2 | `0EA2h`／`0EB0h`／`0ECBh`..`0EEFh` | `STOREVALUE` 收到 `C04B`／`C04C`／`C04D` 時的鏡射 |
| 07 | ECL2 | `1B53h`..`1BA4h` | `MOVEFORWARD`（四個方向，各一組 `inc`／`dec` ＋ 環繞夾回） |
| 11 | INIT | `03F2h`..`03FCh`、`078Eh`..`0798h` | `INITALL`／`INITVARS` |
| 14 | MOVEMENT | `07E6h`..`089Ah`、`0A83h`..`0B07h` | `MOVEPARTY`／`PREMOVEPARTY` |
| 16 | LOADSAVE | `39D1h`／`41DCh` | 存檔／讀檔的 **5 byte 區塊複製** |

⇒ **`INTERPET`（`overlay-02`，擁有 `GOECL` 與 block 載入）一次都沒有寫。**
它在同一份普查裡出現 **23 次讀取**，所以這個零不是掃描沒掃到它。

⇒ **原作沒有「換 ECL block 時的引擎進場放置」。** spec 1172 那句話是錯的。

## 那原作靠什麼決定進新地圖的落點

**腳本自己寫。** 兩個實例：

```
ECL2/0x03（提爾弗頓下水道）
  0042  COMPARE 4C2A 01
  0048  IF =
  0049  GOSUB   80DB
  ...
  00DB  SAVE 00 C04B     ; 進場放置本身就是一段腳本
  00E1  SAVE 00 C04C
  00E7  SAVE 02 C04D
  00ED  CALL 2E10
  00F1  RETURN

ECL3/0x10（猶拉什）
  006C  SAVE 00 C04B
  0072  SAVE 08 C04C
  0078  SAVE 03 C04D
  007E  CALL 2E10
```

下水道那一段還帶著條件（`4C2A == 1`），所以「進場放置」在原作根本不是一個
統一的引擎步驟，而是**每張圖自己決定要不要做、做在哪**。

## `ViewMirror.Block` 因此要重新分類

它不是在模擬引擎行為——原作沒有那個行為。它實際上是**保護 game pack 宣告的
`spawn` 不被腳本自己的進場放置蓋掉**的擋板。

這讓「拿不拿得掉」變成一個**資料問題**而不是模型問題：

- 腳本自己有寫進場座標的地圖 ⇒ 應該讓腳本贏，宣告的 `spawn` 是多餘的。
- 腳本沒寫的地圖 ⇒ 宣告的 `spawn` 是 remake 的補值，必須保留。

⚠ 目前兩者對不上的已知案例：提爾弗頓下水道。腳本說 `(0,0)` 朝南
（`C04D=2` 折成 4），game pack 宣告的 `spawn` 是 `(0,1)`，端到端測試釘的是後者。
差一格，而**還沒有用原版當 oracle 判過誰對**。在判出來之前不動這一格。

## `LOADSAVE` 的 5 byte 區塊：SAVGAM map state 的獨立證人

```
39D1  bf 0f 72       mov  di, 720Fh
39D4  1e 57          push ds ; push di
39D6  b8 05 00       mov  ax, 5
39D9  50             push ax
39E0  9a 33 19 ...   lcall <blockmove>
```

**五個位元組，從 `720Fh` 開始**——正好是 `MapPosX`／`MapPosY`／`MapDirection`／
`MapWallType`／`MapWallRoof`（`720Fh`..`7213h`）。這獨立確認了第 660 輪對
`LoadSAVGAMPrefix` 的修正：那五個位元組就是地圖暫存器本身，不是另一套世界
地圖座標。

## 方法：兩種掃描的失敗方向相反

`cmd/dseg-writers` 是**位元組線性掃描**：會有偽陽性（立即數裡剛好長得像
opcode 的位元組），但對它認得的編碼**不會有偽陰性**。
`cmd/ecl-cell-refs` 走控制流：沒有偽陽性，但掃不到沒被認成程式碼的部分。

問「誰寫 X」的時候偽陰性才是貴的那一種，所以這裡選線性掃描；要下「不存在」
的結論則兩種都要跑。

⚠ 線性掃描仍有兩個盲點，都要靠別的欄位補：

- **段前綴**（`es:`／`ss:`）的存取。
- **執行時算出來的位址**。這一個用「取位址」那一欄補：Turbo Pascal 的全域
  要被指標寫，位址一定得先進暫存器，而那是 `mov r16, imm16` 或 `push imm16`。
  全檔只有 INIT 與 LOADSAVE 取過這三格的位址。

## 正對照

- 掃描器找到 `overlay-07:0EA2h`（`STOREVALUE` 鏡射）與 `1B53h`
  （`MOVEFORWARD`）——兩者都是 spec 1150 已經讀過的已知寫入者。
- 「取位址」找到 `overlay-16:39D1h`，而那一處的下一句就是 `mov ax, 5` ＋
  區塊複製，形狀自證。
- ⚠ 第一版把 `cmpb $0, 0x7211`（`80 3E 11 72 00`）算成寫入，於是 THREED 被列成
  地圖暫存器的寫入者——**那正好是「引擎會自己放置隊伍」需要的證據**。
  `80h`／`81h`／`83h` 的 `/7` 是 `CMP`，它不寫記憶體。
  `cmd/dseg-writers/main_test.go` 用手造位元組把這個分類釘住。

## 明確不宣稱

- 沒有掃 PC-98 版。兩版的 overlay 編號一致（spec 1182），但資料段位移差
  `0x3292`，要查 PC-98 得換一組位址。
- 沒有宣稱下水道正確的落點是 `(0,0)` 還是 `(0,1)`——那要跑原版。
- 沒有宣稱 `MOVEMENT` 那 13 處各自是什麼（只確認它是每步都寫的那一支）。
