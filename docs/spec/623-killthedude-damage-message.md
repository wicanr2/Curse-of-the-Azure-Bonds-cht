# 第六百二十三輪：`KILLTHEDUDE` —— 傷害訊息與硬寫死的 NPC 名字對照表

狀態：`READY`。日期：2026-08-14
位置：PC-98 `overlay-07:2235h`（516 bytes）。

```text
KILLTHEDUDE(char, ?, damage):
    if char^[196h] = 6 then return              ← 已在該狀態就整支跳過
    prompted := 0
    Str(damage:3, num)
    name := copy(char^.name, 30)                ← 名字在紀錄 offset 0
    if char^[0F7h] <> 0 then                    ← 見下節
        if      name = 'ALIAS'           then name := 'エイリアス'
        else if name = 'DRAGONBAIT'      then name := 'ドラゴンベイト'
        else if name = 'AKABAR BEL AKAS' then name := 'アーカバー・ベル・アーカッシュ'
    if char^[1A5h] + 10 < damage then
        msg := '  ' + name + 'は' + '死んだ。'
    else
        msg := '  ' + name + 'は' + num + 'ポイントのダメージを受けた'
    if DS:9637h > 16h then                      ← 文字視窗滿了
        DS:9637h := 11h
        prompted := 1
        <far 0418:12EFh>('リターン・キーを押してください', …, 0, 0Fh)
    DS:9636h := 26h
    <far 0418:0E6Ah>(msg, prompted, 0, 0Fh, 16h, 26h, 11h, 1)
    <far 014A:00ACh>(char, byte(damage))        ← 實際扣血
    <far 0418:14C7h>(0Fh, 26h, 1, 11h)
    if char^[196h] = 5 then char^[196h] := 4
    <far 014A:002Ah>(DS:9594h 所指的紀錄)
```

## 三個 NPC 的名字是**硬寫在這支函式裡**的

字串鏈在 `2192h`，英日成對排列：

| offset | 內容 | offset | 內容 |
|---|---|---|---|
| `2192h` | `ALIAS` | `2198h` | `エイリアス` |
| `21A3h` | `DRAGONBAIT` | `21AEh` | `ドラゴンベイト` |
| `21BDh` | `AKABAR BEL AKAS` | `21CDh` | `アーカバー・ベル・アーカッシュ` |

**紀錄裡存的是英文名**，顯示前才在這裡臨時換掉。判斷方式是**整串字串比對**
（`0A65h:734h`），不是 id 也不是索引——所以名字差一個字就不會被換。

`char^[0F7h] <> 0` 是進入這段的前提。這個旗標非零才做替換，等於「這是劇情
NPC」。

### 對中文化的意含

同一份對照表在中文版要重做一次，而且**不能只改字串內容**：三個日文名的長度
（10／14／30 bytes）與這支函式的 `1Eh`（30）上限綁在一起，中文名超過 30 bytes
會被截斷。另外**比對的左邊是英文原名**，改的是右邊。

這也表示：**其他 NPC 的名字不在這裡**。只有這三個被特別處理，其餘直接用紀錄裡
的內容顯示。

## 角色紀錄欄位

| offset | 內容 | 依據 |
|---|---|---|
| `+0` | 名字（Pascal 短字串，取用上限 30） | `exact` |
| `+0F7h` | 劇情 NPC 旗標（非零才做名字替換） | `exact`（用途為 `strong inference`）|
| `+196h` | 狀態碼：等於 `6` 整支跳過；等於 `5` 則改成 `4` | `exact` |
| `+1A5h` | 目前 HP（byte） | `strong inference` |

`+1A5h` 判定為 HP 的理由是死亡條件本身：

```asm
mov  al, es:[di+1A5h]
xor  ah, ah
add  ax, 0Ah
cmp  ax, [bp+arg_4]      ; 與傷害比
jnb  短：還沒死
```

`傷害 > HP + 10` 才印「死んだ。」——這正是 AD&D 的 **HP 降到 −10 即死亡**。
訊息本身（`N ポイントのダメージを受けた`）也確認 `arg_4` 是傷害值。

注意這**只決定印哪一句訊息**；真正扣血在之後的 `014A:00ACh`。兩者各自判斷，
這裡不假設它們的死亡判定一致。

## 文字視窗游標

`DS:9637h` 超過 `16h`（22）就印「リターン・キーを押してください」並重設為
`11h`（17），`DS:9636h` 固定設成 `26h`（38）。兩者是文字視窗的列與行位置，
`11h..16h` 是可用的訊息區高度。

## 明確不宣稱

- `arg_2`（中間那個參數）的用途——這支函式從頭到尾沒有讀它。
- `+196h` 的完整狀態列舉（只知道 `6` 與 `5→4` 這兩條）。
- `0418h:0E6Ah` 那八個參數各自的意義。
