# 第五百七十九輪：角色記錄的狀態欄位

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`EFFECTS`（overlay-23）。兩平台已配對。

三支狀態轉換函式（`KILLDUDE`／`HEALDUDE`／`STANDUP`）互相印證，釘死了角色
記錄的幾個欄位。

## 欄位

| 偏移 | 內容 | 證據 |
|---|---|---|
| `+78h` | **最大 HP** | `HEALDUDE` 加血後以它為上限 |
| `+0F2h` | effect 鏈頭 far pointer | [spec 578](578-effect-node-list.md) |
| `+196h` | **狀態碼** | 三支都讀寫它；`STANDUP` 設 0＝正常 |
| `+197h` | 旗標（`STANDUP` 設 1、`KILLDUDE` 設 0） | — |
| `+198h` | 類別旗標 | `STANDUP` 用它選訊息；`ATTEMPTTOHIT` 用它選 bank 1 欄位 |
| `+19Ah` | 命中修正 | [spec 577](577-attempttohit-and-effect-chain-walk.md) |
| `+1A5h` | **目前 HP** | `HEALDUDE` 加、`KILLDUDE` 歸零、`STANDUP` 設值 |

## 狀態碼 `+196h`

```text
KILLDUDE(msg, new_state, char)                ← PC-98 0016h／DOS 0016h
    <顯示 msg>
    if char^[196h] in {6, 7, 8} then return   ← 已是這三種就不再改
    char^[196h] := new_state
    char^[197h] := 0
    char^[1A5h] := 0                          ← HP 歸零
    REMOVEFX(char) ; CHECKFX(0Dh, char)
    if char^[197h] = 0 then <far 0189:0013>(char)
    <far 0418:14AA>() ; <far 014A:0089>()
    if DS:7F27h <> 5 then <far 014A:002A>(DS:9594h)

STANDUP(hp, char) -> boolean                  ← PC-98 251Ah／DOS 24EDh
    if not <檢定>(1, …, char) then return false
    char^[196h] := 0 ; char^[197h] := 1 ; char^[1A5h] := hp
    <顯示 (char^[198h] = 1) ? 訊息A : 訊息B>
    return true

HEALDUDE(only_if_hurt, amount, char) -> boolean   ← PC-98 2419h／DOS 23FBh
    if char^[196h] not in {0, 1, 4, 5} then return false
    if only_if_hurt <> 0 and char^[1A5h] >= char^[78h] then return false
    char^[1A5h] := min(char^[1A5h] + amount, char^[78h])
    if char^[197h] = 0 then
        if char^[196h] = 5 then char^[196h] := 4
        if char^[196h] = 4 and DS:7F27h <> 5 then CALLEFFECT(1, nil, char, 4Eh)
    return true
```

已知的狀態碼語意：

| 值 | 依據 |
|---:|---|
| `0` | `STANDUP` 設的值＝可行動 |
| `{0,1,4,5}` | `HEALDUDE` 願意治療的集合（函式前方 32 bytes 的 Turbo Pascal set 常數，兩平台相同） |
| `4` ← `5` | 加血後若 `197h = 0`，`5` 自動降為 `4`，且 `4` 會觸發 `CALLEFFECT(…, 4Eh)` |
| `{6,7,8}` | `KILLDUDE` 視為終局，不再覆寫 |

`HEALDUDE` 的 `only_if_hurt` 為 0 時**滿血也照樣執行**（仍會走 `5→4` 那段），
不是單純的無效呼叫。

## 訊息字串就嵌在 overlay code 段裡

`STANDUP` 的兩則訊息是 **Turbo Pascal 短字串**（長度前綴 ＋ 內容），直接放在
函式前方的 code 段：

| | PC-98 | DOS |
|---|---|---|
| 訊息 A（`198h = 1`） | `は立ち上がり、ニヤリと笑った。`（Shift-JIS，`24EAh`） | 英文 |
| 訊息 B | `は起き上がった。`（`2509h`） | `gets back up` |

**PC-98 版存的是 Shift-JIS，DOS 版存的是英文，兩者長度不同，所以同一則訊息
在兩平台的 overlay-local 位址不同**——不能用固定位移對照，要各自掃描。
掃描與對照見下一輪。

## 明確不宣稱

- `+197h`／`+198h` 的完整語意（只知道被誰讀寫）。
- 狀態碼 `1`／`2`／`3`／`6`／`7`／`8` 各自代表什麼。
- `KILLDUDE` 尾段那四個 far call 的本體。
