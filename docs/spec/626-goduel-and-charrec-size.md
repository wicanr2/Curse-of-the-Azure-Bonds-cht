# 第六百二十六輪：`GODUEL` 複製角色、`SPRIT`／`PIC` 視窗，以及兩平台的 `CHARREC` 大小差一

狀態：`READY`。日期：2026-08-14
位置：PC-98 `overlay-07:1B89h`（584 bytes）、`overlay-07:068Bh`（337 bytes）。
這兩支讀完後 **`overlay-07` 全模組閉合**。

## `CHARREC` 配置大小：DOS `1A6h`、PC-98 `1A7h`

Borland type table 宣告 `CHARREC` 是 `01A7h` = 423 bytes
（[spec 499](499-pc98-alignment-conditional-effects.md)）。但實際配置的數字兩平台
不同：

| 平台 | `New`／`Move` 的參數 | 例 |
|---|---|---|
| PC-98 | `1A7h`（423） | `overlay-07:1C47h`、`overlay-07:1C5Ah` |
| DOS | `1A6h`（422） | `overlay-02:05C9h`（GetMem）、`overlay-02:05FFh`（Move）|

`mov ax, 1A7h`（位元組 `B8 A7 01`）在**全部 DOS overlay 裡出現 0 次**——不是抽樣沒
抽到，是整份掃過都沒有。

### 但實際佔用一樣

DOS 的 `New` wrapper 會把大小對齊到 8 bytes（`(size + 7) and 0FFF8h`）：

| 宣告 | 對齊後 |
|---|---|
| `1A6h`（422） | **424** |
| `1A7h`（423） | **424** |

兩邊都落在 424。所以這個差一**不會**造成 DOS 版少配一個 byte——最後那個
byte（`+1A6h`，由 ECL 特殊位址 `7D1Bh` 讀取，見
[spec 624](624-ecl-special-address-space.md)）在兩個平台都在配置範圍內。

remake 若自己配置 423 bytes 也不會有問題；但**`Move` 的長度要照平台**，DOS 版複製
角色時最後一個 byte 不會被搬過去。

### 欄位偏移是共用的

spec 499 由 DOS type table 定出 `+119h` = `GENDER`、`+11Bh` = `ALIGNMENT`，而
[spec 624](624-ecl-special-address-space.md) 在 PC-98 的 `CHECKSPECIALS` 讀到的正是
`+119h` 與 `+11Bh`，兩者都用 `cbw`（有號）。**欄位配置兩平台相同**，差的只有配置
長度那一個數字。

## `GODUEL`（`1B89h`）

```text
GODUEL(mode):
    DS:0BDE8h := 1
    bank1^[5CCh] := signext(mode)
    me := DS:9594h
    p := DS:9598h                          ← 角色鏈
    while p <> nil do
        if p.name <> me.name then p^[197h] := 0    ← 比名字，不是比指標
        p := p^[18Ah]
    if mode = 0 then return

    <sub_1979>(DS:0A895h, 0Bh)             ← 資源名 'CPIC'
    last := 走到角色鏈尾端
    New(clone, 1A7h)
    Move(me^, clone^, 1A7h)                ← 整個複製一份
    clone^[197h] := 1
    clone^[18Ah] := nil
    clone.name := 'ロルフ'                  ← 上限 15
    clone^[199h] := 1 ; clone^[198h] := 1
    clone^[0F7h] := 0B2h                   ← 劇情 NPC 旗標
    clone^[143h] := DS:0A895h
    clone^[0F2h] := 0 ; clone^[0F4h] := 0   ← 清掉 effect 鏈
    clone^[14Eh] := nil                     ← 清掉物品鏈
    last^[18Ah] := clone                    ← 接在尾端

    src := me^[14Eh]                        ← 再逐件複製物品
    while src <> nil do
        old := clone^[14Eh]
        New(clone^[14Eh], 67h)
        Move(src^, clone^[14Eh]^, 67h)
        clone^[14Eh]^[52h] := old           ← 插在頭部（順序反過來）
        src := src^[52h]
        Dispose(clone^[14Eh], 67h)          ← ⚠ 見下
```

複製出來的角色叫**ロルフ**（Rolf），`+0F7h` 設成 `0B2h`。這個值正好是
[spec 624](624-ecl-special-address-space.md) 裡 `STORESPECIALS` 寫 `+0F7h` 時的門檻
（`if value > 0B2h then value := value − 32h`）——`0B2h` 本身不會被減。

比對用的是**名字字串**（`0A65h:734h` 作用在紀錄 offset 0），不是指標。同名的兩筆
會被當成同一個。

### 物品複製迴圈每一輪都把剛建好的節點釋放掉

迴圈末端對 `@clone^[14Eh]` 呼叫 `081F:01A2h`，參數大小 `67h`，正是這一輪剛
`New` 出來的那個節點。

`081F:01A2h` 是 `Dispose`／`FreeMem`：它與配置用的 `081F:0043h` 在全 PC-98 overlay
的呼叫大小分佈完全對齊（`1A7h` 各 7 次、`1Eh` 各 4 次、`11Dh` 各 2 次、
`0AC8h`／`0BCh`／`4E9h` 各 1 次）。等級 `strong inference`——不是靠名稱，是靠
配置與釋放必然成對出現的大小簽章。

後果：迴圈結束時 `clone^[14Eh]` 指向已釋放的記憶體，而且每一輪都會走「已有物品」
那條分支（指標非 nil），把上一輪的野指標接成 next。**這是原作行為**，不是判讀
錯誤；remake 要照抄的話得連這個一起抄，或明確記錄為修正項。

⚠ 這條需要 runtime 驗證再下最終結論。靜態證據只到「`01A2h` 是釋放」為止。

## `068Bh`：`SPRIT` 與 `PIC` 兩種資源

```text
(state, ?, arg_4, arg_6, arg_8):
    saved := DS:0A2A8h
    if state^[1] = 0 then
        if DS:0A2A8h <> 0 then DS:0A2A8h := 0 ; DS:0A66Ch := 1 ; <sub_1745>()
        if state^[0] = 0 then
            if bank0^[1CCh] <> 0 then
                <sub_1794>(DS:0A2C6h, 1, arg_8)    ← 資源名 'SPRIT'
                state^[0] := 1 ; DS:0BDF7h := 1
        else
            DS:0A66Ch := 1 ; <sub_1745>()
        if DS:7F27h = 4 then <sub_17A8>(DS:0A2C6h, arg_4 + 1)
    …
    if bank1^[5C2h] = 0FFh then
        <sub_1794>(DS:0A2C6h, 0, arg_6)            ← 資源名 'PIC'
        state^[1] := 1
        <far 0176:0027>(DS:0A2CDh, DS:0A2CFh, 1, 3, 3)
    else
        <sub_64C>(arg_6, bank1^[5C2h])
        state^[1] := 1 ; DS:0BDF5h := 0
    DS:0A2A8h := saved                              ← 進出成對
```

三個資源名以 Pascal 短字串內嵌在 code 段：`0681h` = `SPRIT`、`0687h` = `PIC`、
`1B7Dh` = `CPIC`。`sub_1794` 的第二個參數 `1`／`0` 就是選哪一種。

`state` 是一個 2-byte 記錄：`[0]` 表示 sprite 已載入、`[1]` 表示圖片已載入。
`DS:0A2A8h` 在函式頭尾成對存回，中途可能被清成 0。

`bank1^[5C2h] = 0FFh` 代表「沒有指定替代圖」；非 `0FFh` 時改走 `sub_64C`。

## 明確不宣稱

- `GODUEL` 的 `mode` 除了「0 就只清旗標」以外的語意。
- `+197h`／`+198h`／`+199h`／`+143h` 各自代表什麼。
- `sub_1745`／`sub_1794`／`sub_17A8`／`sub_1979` 的內部行為（本輪只確定呼叫形狀與
  傳入的資源名）。
- PC-98 的 `New` wrapper 是否同樣對齊到 8 bytes（本輪只驗證 DOS 側）。
