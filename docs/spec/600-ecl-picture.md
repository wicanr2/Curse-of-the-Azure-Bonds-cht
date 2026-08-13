# 第六百輪：`0Eh`（顯示圖片）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:08A7h`（235 bytes）。

```text
READVAR(1)
v := ADDRESSVALUE(1)
if v = FFh then                               ← FFh ＝ 關閉
    if (DS:7F28h = 4 and DS:7F27h = 4) or DS:7F28h <> 4 then
        if DS:BDF4h <> 0 or DS:BDF7h <> 0 then
            DS:A66Ch := 1 ; <far 0172:0025>()
            DS:BDF4h := 0 ; DS:BDF7h := 0 ; DS:BDF5h := 1
    DS:BDDAh := 0 ; DS:BDDBh := 0
else
    DS:BDDBh := 1 ; DS:BDF4h := 1
    if bank1^[5C2h] = FFh then
        DS:BDF5h := 1
        if v >= 78h then
            <far 0176:004D>(v) ; <far 0176:0052>() ; DS:A66Ch := 0
        else
            s := 'PIC'                        ← 檔名前綴，字串常數就在函式前面
            <far 0176:0034>(DS:A2C6h, 0, v)
            <far 0176:0025>(DS:A2CDh, 1, 3, 3)
    else
        <far 0062:0043>(v, bank1^[5C2h])
```

三件事：

1. **`FFh` 是「關閉圖片」的哨兵值**，不是編號 255。這與 `21h`／`37h` 用 `FFh`
   表示「這個 operand 沒有值」（[spec 587](587-ecl-handler-21-37-shared.md)）
   一致——`FFh` 在這個引擎裡普遍當哨兵。
2. **編號 `78h` 是分界**：`>= 78h` 走一條路（`0176:004D`），`< 78h` 走另一條
   （帶著 `'PIC'` 這個檔名前綴）。所以低編號是**從 `PIC*` 檔載入**、高編號是
   另一種來源。
3. **`bank1^[5C2h] <> FFh` 時整個跳過**，改呼叫 `0062:0043`。那是另一種顯示
   模式（戰鬥中？），同一條 ECL 指令的效果因此隨狀態而異。

## 明確不宣稱

- `78h` 之上那條路載入的是什麼。
- `DS:BDDAh`／`BDDBh`／`BDF4h`／`BDF5h`／`BDF7h`／`A66Ch` 各自的語意——只確定
  它們是這條指令與 `31h`（[spec 586](586-ecl-handlers-31-33-34.md)）共用的
  一組旗標。
- bank 1 `+5C2h` 的語意。
