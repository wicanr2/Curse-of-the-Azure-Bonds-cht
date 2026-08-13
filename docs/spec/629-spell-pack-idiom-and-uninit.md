# 第六百二十九輪：法術三支 —— 「高 4 bit 打包」的固定寫法與一個未初始化分支

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-22:2690h`（108 bytes）、`29A3h`（97）、`2776h`（183）。

## 打包成一個 byte 的固定寫法

```asm
call far 013Eh:004Dh        ; ROLLDICE
xor  ah, ah
mov  cx, 4
shl  ax, cl                 ; 結果放進高 4 bit
mov  [bp+var_1], al
...
call far 014Ah:00D4h
or   al, [bp+var_1]         ; 另一個值放低 4 bit
```

`29A3h` 與 [spec 628](628-spell-effect-wrappers.md) 的 `3804h` 用同一套動作：
**`shl` 之後 `or`／`add`，是把兩個值塞進同一個字，不是乘法也不是相加**。
高位是「數量」，低位來自 `014Ah:00D4h`。

**移位量隨呼叫端不同**：`29A3h` 與 `3804h` 是 `shl 4`，
[spec 630](630-spell-target-array.md) 的 `2147h` 是 `shl 7`。所以這不是固定的
「高低 nibble」，是欄位寬度視內容而定的位元打包。

`29A3h` 的數量是 `ROLLDICE(1, 4)` ＝ `1d4`，訊息是「は分身した。」。`1d4` 個分身
與 AD&D 的 **Mirror Image** 一致（`strong inference`，本輪不寫進結論）。

注意 `014Ah:00D4h` 在 `29A3h` 只推一個參數、在 `3804h` 推兩個。呼叫形狀不同，
本輪不宣稱它的參數表。

## `2690h`：`DS:0A520h` 不在 1..4 時讀到未初始化的值

```text
case DS:0A520h of
    1: if DS:0A031h = 17h then v := 0FEh else v := 0FDh
    2: v := 0FFh
    3, 4: v := 0
end                                ← 沒有 else
<sub_1D0B>(v, 'は金縛りにあった。')
```

`v`（`[bp+var_1]`）**只在四個 case 裡被賦值**，函式開頭沒有預設值。`DS:0A520h`
落在 `1..4` 以外時傳給 `sub_1D0B` 的是堆疊殘值。

這與 [spec 624](624-ecl-special-address-space.md) 記到的 `CHECKSPECIALS` `7D0Dh`
是同一類原作行為——**兩處都是「case 沒有 else」**，不是同一支函式的問題。
remake 照抄的話要一併決定怎麼處理，或明確記錄為修正項。

`v` 的四個值 `0FDh`／`0FEh`／`0FFh`／`0` 看起來是有號的 `−3`／`−2`／`−1`／`0`，
但 `sub_1D0B` 怎麼用它本輪沒讀，所以不下結論。

## `2776h`：目標無效就把指標清成 nil

```text
t := DS:0A521h                                  ← far pointer
if t^[196h] = 1 then
    DS:0A521h := nil                            ← 直接清掉，不印任何訊息
else if <far 014Ah:00A7h>(t, 37h, @var_4) <> 0 then
    if t^[1A5h] = 0 then t^[1A5h] := 1          ← HP 為 0 就拉到 1
    推入 DS:0A031h、0FFh、1、0、0，備妥 'は魔法にかかった。'，<sub_F62>
    <far 013Eh:002Ah>(4Eh, t, 0, 1)
```

`+1A5h` 是目前 HP（[spec 623](623-killthedude-damage-message.md) 由死亡門檻定
出）。**`HP = 0` 才拉到 1，不是治療到滿**——這支法術不會讓活著的人加血。

`t^[196h] = 1` 這條路徑**把全域的目標指標清成 nil 就結束**，沒有訊息。呼叫端若
不檢查，後面會拿到 nil。

## 明確不宣稱

- `sub_1D0B`／`sub_F62`／`014Ah:00A7h`／`013Eh:002Ah`／`014Ah:00D4h` 的內部行為。
- 效果 id `37h`、常數 `4Eh`、`DS:0A520h`／`DS:0A031h` 的身分。
- 這三支各自對應哪一個法術名稱。
