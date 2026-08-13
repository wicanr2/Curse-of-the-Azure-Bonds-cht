# 第六百四十一輪：兩平台欄位偏移對照表（由 190 對配對函式量出）

狀態：`READY`。等級：`exact`。日期：2026-08-14
工具：`scripts/field_offset_map.py`；資料：`docs/audit/cross-platform-field-offsets.json`
位置：DOS `overlay-12` 的 `1C1Ch`、`1C4Fh`、`1C90h`、`1FE4h`、`203Ch`、`20FAh`。

## 量法

兩平台是同一份 Turbo Pascal 原始碼編譯的，同一支程序的**指令種類與順序**一致，
只有運算元不同。所以：取助憶碼序列在各自模組內唯一、長度 `>= 12` 的配對函式，把
兩邊的 `[di+XXXh]` **依序**對起來，就直接讀出對應關係——不必先知道欄位是什麼，
也不必猜。

全模組掃出 **190 對**配對，多數偏移的票數是滿票（`+18Eh` 106/107、`+18Ah` 56/56、
`+52h` 49/49）。

## 角色紀錄（`CHARREC`）

| PC-98 偏移 | DOS 偏移 | 差 |
|---|---|---|
| `+00h` .. `+131h` | **相同** | 0 |
| `+14Ch` .. `+1A6h` | **PC-98 − 1** | 1 |

[spec 626](626-goduel-and-charrec-size.md) 量到配置大小差 1（DOS `1A6h`、PC-98
`1A7h`）。這裡看到**那一個 byte 插在 `+131h` 與 `+14Ch` 之間**。

實測到的邊界兩側：

```text
pc98 +131h ↔ dos +131h    （2 票，差 0）
pc98 +14Ch ↔ dos +14Bh    （2 票，差 1）
```

中間 `+132h`..`+14Bh` 這段沒有配對函式碰到，所以**確切的插入位置還沒收斂**，只知
道落在這 26 bytes 之內。

### 幾個已命名欄位的兩平台位置

| 欄位 | PC-98 | DOS | 出處 |
|---|---|---|---|
| `RACETYPE` | `+11Ah` | `+11Ah` | [spec 499](499-pc98-alignment-conditional-effects.md) |
| `ALIGNMENT` | `+11Bh` | `+11Bh` | 同上 |
| 五種硬幣 | `+0FBh`..`+103h` | 相同 | [spec 622](622-character-money-block.md) |
| 劇情 NPC 旗標 | `+0F7h` | `+0F7h` | [spec 623](623-killthedude-damage-message.md) |
| effect 鏈頭 | `+0F2h` | `+0F2h` | [spec 604](604-ecl-spawn-monsters.md) |
| 物品鏈頭 | `+14Eh` | **`+14Dh`** | [spec 621](621-ecl2-robstuff.md) |
| next 指標 | `+18Ah` | **`+189h`** | [spec 625](625-ecl-address-class-and-bank-wrap.md) |
| 狀態碼 | `+196h` | **`+195h`** | [spec 623](623-killthedude-damage-message.md) |
| 目前 HP | `+1A5h` | **`+1A4h`** | 同上 |

**所有 `+14Ch` 以後的欄位，PC-98 的判讀套到 DOS 都要減 1。**

## 物品節點：差 40（`28h`）

```text
pc98 +51h..+66h  ↔  dos +29h..+3Eh      整段固定差 28h
```

`+52h`（物品鏈的 next 指標，49/49 票）在 DOS 是 `+2Ah`。這**不是差 1 那種微調**，
是整個結構的欄位位置差 40 bytes——PC-98 版在物品節點前段多了 40 bytes 的東西。

節點大小本身：PC-98 是 `67h`（103，[spec 610](610-ecl-load-encounter-items.md)）。
DOS 側的對應大小本輪未量。

## 六支的內容

```text
1C1Ch:  if DS:6F95h and 4 <> 0 then
            <sub_1B>(0)
            <far 013Eh:007Dh>(8, 1)          ← 回傳值存進區域變數後沒再用到
            arg_6^[1A4h] := arg_6^[1A4h] + 8     ← 目前 HP，+8 且不夾上限

1C4Fh:  p := <sub_FAA>(DS:6506h)
        if p <> nil and byte[p^[2Eh] × 16 + 5CFDh] = 1 then
            DS:6F94h := 1

1C90h:  if DS:758Dh = 0 and DS:6F9Fh = 0 then
            arg_2^[3] := arg_2^[3] and 0Fh              ← 清掉高 nibble
        else if arg_2^[3] and 10h = 0 then
            DS:6F9Fh := 0FFh
            arg_2^[3] := arg_2^[3] or 10h               ← 設 bit 4

1FE4h:  p := <sub_FAA>(DS:6506h)
        if p <> nil and (byte[p^[2Eh] × 16 + 5CFDh] and 81h) <> 0 then
            DS:6F94h := DS:6F94h div 2                  ← 有號

203Ch:  arg_2^[4] := 0
        if arg_6^[196h] <> 0 then                       ← ＝ PC-98 的 +197h
            備妥 'Falls dead'，<far 013Eh:0005h>(arg_6, 訊息, 6)

20FAh:  arg_6^[1A4h] := min(arg_6^[1A4h] + 3, arg_6^[78h])
```

`20FAh` 把目前 HP 加 3 後夾在 `+78h`——`+78h` 兩平台相同（14/14 票），是上限。
對照 `1C1Ch` 的 `+8` **沒有夾**：同一個欄位，兩支的處理不一樣。

`203Ch` 讀的 DOS `+196h` 對應 PC-98 的 `+197h`，**不是**狀態碼（狀態碼在 DOS 是
`+195h`）。

## 全域變數的兩平台對應

| 用途 | DOS | PC-98 |
|---|---|---|
| 傷害值 | `DS:6F94h` | `DS:0A02Eh`（[spec 640](640-save-for-half-and-damage-global.md)）|
| 旗標欄位 | `DS:6F95h` | `DS:0A02Fh` |
| 16 bytes 一筆的表 | `DS:5CFDh` | `DS:61BCh` |

依據是**動作相同**：`1FE4h` 與 PC-98 的 `1A34h` 都是「查 16 bytes 表 → 對傷害值做
有號除以 2」；`1C1Ch` 與 PC-98 的 `176Dh` 都是「查旗標的某個 bit → 呼叫
`sub_1B(0)`」。`field_offset_map.py` 也獨立量到
`pc98 +61BCh ↔ dos +37E2h`（差 `2AECh`），與這個對應一致。

## 明確不宣稱

- 多出來的那個 byte 確切插在 `+132h`..`+14Bh` 的哪一格。
- 物品節點為什麼差 40 bytes、DOS 側的節點大小。
- `+78h`／`+2Eh`／`+3`／`+4` 欄位的意義。
- `DS:758Dh`／`DS:6F9Fh`／`DS:6506h` 的身分。
