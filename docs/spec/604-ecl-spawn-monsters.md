# 第六百零四輪：`0Bh`（產生 N 個角色／怪物）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:049Ch..0857h`（**952 bytes**，328 條指令；IDA 標 452）。

```text
saved := DS:9594h                             ← 目前目標，離開時還原
READVAR(3)
if DS:789Ch >= 3Fh then goto done             ← 名額上限 63
<顯示>「読み込み中です」（`unk_488`）
<載入範本>(ADDRESSVALUE(1), @template, @items, @effects)
Move(template^, @rec, 1A7h)                   ← 記錄 423 bytes
count := ADDRESSVALUE(2) ; if count = 0 then count := 1
kind  := ADDRESSVALUE(3)
<…>('CPIC', DS:A895h, kind)                   ← `unk_497` 是圖檔前綴
<走到 DS:9598h 鏈尾，接上第一筆>
repeat
    if DS:789Ch >= 3Fh then break
    New(node^[18Ah], 1A7h) ; Move(@rec, 新節點^, 1A7h)
    新節點^[143h] := DS:A895h
    新節點^[18Ah] := nil
    新節點^[0F2h] := nil ; 新節點^[14Eh] := nil    ← 兩條鏈先清空
    DS:789Ch := DS:789Ch + 1
    for each item in 範本物品鏈 do
        New(節點^[14Eh], 67h) ; Move(item, …, 67h)
    for each fx in 範本 effect 鏈 do
        New(節點^[0F2h], 9) ; Move(fx, …, 9)
until 已產生 count 個
DS:A895h := DS:A895h + 1
DS:BDFBh := 1
done:
DS:9594h := saved
```

## 記錄大小是 423 bytes

`New` 與 `Move` 的參數直接給出 **`1A7h` ＝ 423**。先前只從欄位偏移知道至少到
`+1A6h`（[spec 603](603-ecl-party-stat-aggregate.md)），現在確認總長，
`+1A6h` 正好是最後一個 byte——自洽。

三種節點的大小因此都確定了：

| 節點 | 大小 | 出處 |
|---|---:|---|
| 角色／怪物記錄 | `1A7h`（423） | 本輪 |
| 物品 | `67h`（103） | [spec 596](596-ecl-party-item-sweep.md) |
| effect | `9` | [spec 578](578-effect-node-list.md) |

## 複製鏈時**插在鏈頭**——順序會反轉

複製範本的物品與 effect 時，新節點是**插在鏈頭**（新節點的 next 指向舊鏈頭），
所以複製出來的鏈**順序與範本相反**。

這與 `ADDEFFECT`（[spec 578](578-effect-node-list.md)）把新效果**接在鏈尾**
不同。同一個引擎裡兩種插法都有，remake 兩邊都要照抄——effect 的遍歷順序會
影響哪一個先生效。

## 其他

- **名額上限 63**（`DS:789Ch`），而且**迴圈中每次都重新檢查**：要求 5 個但只
  剩 2 個名額時，會產生 2 個然後停，不是全部放棄。
- `DS:A895h` 是群組編號，寫進每個新記錄的 `+143h`，整批做完才加一。
- `DS:9594h`（目前目標）進來存、離開還原——這條指令不改變目標。

## 明確不宣稱

- `<載入範本>`（`0062:0039`）從哪裡讀範本。
- `DS:789Ch`／`DS:A895h`／`DS:BDFBh` 除了上述用途之外的語意。
- 記錄 `+143h` 的用途（只確定存的是群組編號）。
