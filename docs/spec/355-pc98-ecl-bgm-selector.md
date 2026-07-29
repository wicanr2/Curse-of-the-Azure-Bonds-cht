# 355：PC-98 ECL 區塊與 BGM selector

狀態：`READY`（selector 分支與資料流）；曲名及實際音訊解碼仍為
`IN PROGRESS`

## 範圍

本規格只回答 PC-9801 版何時向音樂 wrapper 傳入哪個 selector。它不把網路
播放清單的曲名順序直接套到 driver index，也不宣稱缺少的 `MSCDRV` 媒體已
解碼或可播放。

## 地址校準

Borland debug symbol、IDA data segment 與指令 bytes 交叉驗證如下：

- IDA data segment 線性起點：`0x1C290`。
- Borland runtime data segment：`0C29`。
- `WLDTWN`：`0C29:7F11`，對應 IDA `byte_241A1`。
- `CURRENTECL`：`0C29:BDF0`，對應 IDA `byte_28080`。
- `MSCPLAY`：`0893:0114`。
- `BGMPLAY`：`0893:0177`，對應 IDA `sub_18AA7`。

因此 BGMPLAY 讀取的不是未命名「area code」，而是目前全域 ECL block ID。

## Writer／consumer 證據

### Writer

TPOV overlay 2（compiler module `INTERPET`）在 local code `0x0C26` 附近：

1. 把舊 `CURRENTECL` 保存至 PARTY 結構 `+0x1E4`。
2. 呼叫 ECL loader selector。
3. 把回傳的 `AL` 寫入 `ds:BDF0`，也就是 `CURRENTECL`。
4. 使用該值載入目的 ECL。

同一 overlay 在 `0x3A77` 附近把預設值 `1` 寫入 `CURRENTECL`，或由 PARTY
結構恢復先前值。這證明它是跨 loader／party transition 保存的 ECL 身份。

### Consumer

完整 far-call bytes：

```text
9A 77 01 93 08
```

在 `GAME.EXE`／`GAME.OVR` 的 TPOV code 中只出現一次，位於 overlay 26
local `0x0160`；反組譯為 `call 0893:0177`。overlay/module 序號關係與
`INTRO` 實例支持它位於 `MENUS`，但此 module 歸屬目前標為
`nearby`，不影響唯一 callsite 與 selector 分支的 `exact` 結論。

`BGMPLAY` 無參數，直接讀 `CURRENTECL`。它把下列 1-based selector 傳給
`MSCPLAY`；`MSCPLAY` 再減一，才交給 driver。

## 精確 selector 表

| CURRENTECL | BGMPLAY selector | driver index | 信心 |
|---|---:|---:|---|
| `01`, `31` | 3 | 2 | exact |
| `11`, `12`, `21`, `22`, `23`, `15`, `43`, `45` | 4 | 3 | exact |
| `50`, `51` 且 `WLDTWN == 0` | 5 | 4 | exact |
| `50`, `51` 且 `WLDTWN != 0` | 6 | 5 | exact |
| `20`, `40`, `42` | 8 | 7 | exact |
| `02`, `10`, `05`, `35` | 9 | 8 | exact |
| `03`, `04`, `25`, `32`, `33` | 12 | 11 | exact |
| `30` | 不變 | 不變 | exact |
| `52` | 無分支 | 無分支 | exact |

原始 switch 的較晚分支也列出 `23 → 8`，但 `23` 已被較早的
`23 → 4` 命中，因此實際可達結果是 selector 4。`05` 存在於 switch，但目前
實際 ECL corpus 沒有 block `05`；game-pack 保留它，以忠實描述 executable。

## ECL namespace 交叉驗證

DOS DAX inventory 與既有 spec 198／273 證明 25 個實際 block：

- ECL2：`01`, `02`, `03`, `04`
- ECL3：`10`, `11`, `12`, `15`
- ECL4：`20`, `21`, `22`, `23`, `25`
- ECL5：`30`, `31`, `32`, `33`, `35`
- ECL6：`40`, `42`, `43`, `45`
- ECL1：`50`, `51`, `52`

這與 BGMPLAY 幾乎覆蓋全部非 demo block 的形狀一致，進一步支持
`CURRENTECL` 校準。

## Remake contract

作品中立 engine 提供：

- `music_tracks`：穩定 track ID、來源平台、wrapper selector、driver index。
- `music_bindings`：ECL block、可選 opaque context、track ID。
- `music_cues`：ECL signal raw value、opaque context。
- `FindMusicBinding`：先找 exact context，再找 context-free fallback。

CoAB game-pack 使用 `pc98-bgm-selector-*` 作穩定 ID。曲名在音序列或 runtime
聲音被交叉驗證前不得加入；`WLDTWN` 的兩支以
第 362 輪已由 IDA writer、PC-98 decoded ECL 與 DOS 同場景英文流程三方
證明：`WLDTWN == 0` 是區域／戶外導航，非零是
`pc98-town-services-menu`（城鎮設施選單）。`0x50/0x51` 初始 binding
可直接使用 selector 5；selector 6 保留 exact context。完整證據見
[`362-pc98-disk-b-dax-and-wldtwn.md`](362-pc98-disk-b-dax-and-wldtwn.md)。

遊戲規則層在初始 ECL 與 NEWECL block 改變時產生一次 `MusicEvent`。事件
也會依 `PICTURE 0x50/0x79` cue 在同 block 內切換 selector 6／5，並抑制
同曲重播。block `0x50/0x51` 的正常 ECL 玩家路徑已驗證。事件不依賴
Ebiten 或音訊裝置；實際播放 adapter 與音樂資產解碼屬下一階段。

## 待完成

- 補齊 PC-98 `MSCDRV` 缺少的 1 KiB 或由 runtime capture 取得完整 driver
  input/output。
- 依
  [`364-pc98-music-vector-bridge.md`](364-pc98-music-vector-bridge.md)
  的 READY 7Eh／D2h ABI，繼續命名低階 D2h dispatch 與 `CEE0` provider。
- 對每個 selector 錄製可重現音訊，與公開播放清單及遊戲場景交叉驗證。
- 實作循環、切換、停止及存檔恢復語意，再接上跨平台播放 adapter。
