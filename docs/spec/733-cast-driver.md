# 第七百三十三輪：施法主流程，與 `DS:7563h` 的身分

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-22:1263h`（266 條指令）。

## `DS:7563h` 區分「法術」與「物品」

同一支函式裡三處成對出現：

```text
if DS:7563h = 0 then                       ← 法術
    顯示法術名 ＋ "can't be cast here..."
    問 'Lose it? '
else                                        ← 物品
    顯示 'That Item' ＋ 'is a combat-only item...'
    問 'Use it? '
```

`'That Item'`／`'combat-only item'` 只出現在 `DS:7563h <> 0` 那一側，所以

> **`DS:7563h <> 0` 代表這次施放來自物品，不是角色記憶的法術。**

這正好解掉 spec 705 的未定項：`0D67h` 在 `DS:7563h <> 0` 時把等級**寫死成 6**
去算持續時間。也就是**物品施放一律當 6 級**，不看使用者的等級。

## 主流程

```text
p  := DS:6506h
ok := 1

if DS:4FBAh <> 5 then                                  ← 不在戰鬥畫面
    if 表[法術編號].+7 = 0 then                         ← 選目標模式 0（spec 719 的「其他」）
        法術：畫法術名 ＋ "can't be cast here..."，問 'Lose it? '
              → 'Y' 就 <overlay-24 entry#16>(法術編號, p)
        物品：畫 'That Item' ＋ 'is a combat-only item...'，問 'Use it? '
              → 'Y' 就 arg_2^ := 1
        arg_6 := 0；ok := 0

if <overlay-24 entry#27>(p, 4Ah, @node) 找到 then       ← 身上有效果 4Ah
    if ROLLDICE(1, 2) = 1 then                          ← 五成
        <sub_6209h>(法術編號, 'miscasts', p)
        arg_6 := 0；ok := 0

if arg_6 <> 0 且 DS:7563h = 0 then
    <sub_6209h>(法術編號, 'casts', p)

if ok = 0 then 跳到收尾

<[DS:72A0h] 第 0 項>(arg_2, arg_8, 法術編號)             ← 選目標
if arg_2^ = 0 then 跳到「取消」

（戰鬥畫面才做）走 overlay-31 entry#4 找一個可用的動畫槽、
overlay-32 entry#14／#12 播放，音效依 arg_8 是 2Fh／33h／其他選三種之一

DS:6F97h := 法術編號
call dword ptr [DS:72A0h + 法術編號 × 4]                ← 分派（spec 721）
DS:6F97h := 0；DS:6F9Dh := 0

「取消」：
    if DS:4FBAh <> 5 then ok := 0
    else
        if arg_8 = 0 then 問 'Abort Spell? '，不是 'Y' 就回到選目標
        顯示 'Spell Aborted'
        法術（`DS:7563h = 0`）才 <overlay-24 entry#16>(法術編號, p)
        ok := 0
```

## 三件值得記的

**「不能在這裡施放」的判準是選目標模式為 0。** spec 719 讀 `10D2h` 時只看到
模式 `1`／`2`／`4` 有處理、其他一律失敗；這裡補上了另一半：**模式 0 就是
「只能在戰鬥中用」的標記**，而且會在非戰鬥畫面被攔下來並問玩家要不要放棄。

**`miscasts` 是五成機率。** 條件是身上有效果 `4Ah`，然後 `ROLLDICE(1, 2) = 1`。
不受等級或屬性影響。

**取消也會消耗法術。** `Spell Aborted` 之後仍然呼叫 `<overlay-24 entry#16>`
（和 `Lose it?` 回答 `Y` 是同一支），只有物品那一側跳過。所以在戰鬥中選完目標
又反悔，法術照樣沒了。

## 用到的字串

| 位址 | 內容 |
|---|---|
| `CS:11ECh` | `can't be cast here...` |
| `CS:1202h` | `Lose it? ` |
| `CS:120Ch` | `That Item` |
| `CS:1216h` | `is a combat-only item...` |
| `CS:122Fh` | `Use it? ` |
| `CS:1238h` | `miscasts` |
| `CS:1241h` | `casts` |
| `CS:1247h` | `Abort Spell? ` |
| `CS:1255h` | `Spell Aborted` |

`Lose it? `／`Use it? `／`Abort Spell? ` 三個都以**空白結尾**——游標會接在問號後
一格。中文化時那個尾隨空白要留著，否則輸入位置會貼著問號。

Y／N 問答走 `overlay-26 entry#6`，比對的是 `'Y'`（`59h`）。**按鍵是硬比對大寫
`Y`**，所以中文化不能只改提示文字而不動判斷。

## 明確不宣稱

- 效果編號 `4Ah` 的名稱（只知道它會導致 `miscasts`）。
- `overlay-24 entry#16` 做什麼（推測是扣掉已記憶的法術）。
- 音效編號 `2Fh`／`33h` 的差別。
