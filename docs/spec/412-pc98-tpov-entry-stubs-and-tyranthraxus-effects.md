# 第四百一十二輪：PC-98 TPOV entry stub 與提朗瑟克斯效果表（READY）

## 範圍與結論

本輪以唯讀 PC-98 磁碟映像、raw bytes、Borland symbols 與 Docker 內
IDA Pro 9.4，關閉 resident far pointer 到 overlay-local handler 的位址橋接。
這證明提朗瑟克斯六筆 `MON6SPC` 的靜態 handler 與基本語意；不代表閃電術
AI、目標、地形反彈、動畫、音效或全部傷害 boundary 已完成。

## 輸入與非破壞性

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | resident controls、entry stubs、Borland symbols | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | 36 段 code 與 fixup corpus | `exact` |
| overlay 12 code | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` | EFFPROCS handlers | `exact` |

原始 `pc98-disk1.raw` 全程唯讀；抽出的 executable、overlay、IDA database 與
報告只存在 `/tmp` 副本。IDC 只附加指令邊界與報告標籤，沒有改寫原始檔、
原始位置或 symbols。

## TPOV typed grammar

PC-98 MZ header 是 `9Ah` paragraphs，load image 從 file `09A0h` 開始。
overlay 12 control 在 file `1250h`，故 resident load-image offset 是 `08B0h`，
即 segment `008Bh`。control 固定頭為 `20h` bytes，後接 `EntryCount` 筆：

```text
entry_stub := CD 3F, handler_local_offset:u16le, flags:u8
stub_offset = 20h + entry_index * 5
```

overlay code 後的 relocation span 是嚴格遞增的 `u16le` code fixup offsets；
每一筆必須容納完整兩 byte segment fixup。它不是可反組譯 code。36 段 control
chain 全部通過 entry signature、handler bound、fixup bound 與遞增驗證；
overlay 12 有 136 entries、361 fixups。

resolver 必須同時匹配 resident segment 與 stub offset，不能只因數字相同就
跨位址空間合併：

```text
resident segment:stub offset
  -> control segment + entry index
  -> entry.handler_local_offset
  -> 指定 overlay 的 code-local handler
```

## 六筆效果的 exact 對應

| effect | resident pointer／來源 | entry → handler | 靜態語意 | 等級 |
|---|---|---|---|---|
| `18h` | `008B:02C3` | 135 → `2ECBh` | 偵測隱形；handler 本身為 no-op，消費端另見 spec 410 | `exact` |
| `4Fh` | `008B:01A6` | 78 → `19B3h` | 設 Fire＋Magic flags，擲 `2d10` 並進入傷害流程 | `exact` |
| `6Ah` | `008B:0214` | 100 → `2404h` | 15% base 魔法抗性 wrapper | `exact` |
| `70h` | `008B:0232` | 106 → `249Dh` | damage flags 含 Fire 時呼叫 `Protected(0)` | `exact` |
| `84h` | overlay 22 `A250/A252 ← 0117:025F` | 115 → `62D7h` | 呼叫 spell `33h`，即 Lightning Bolt | `exact` |
| `87h` | `008B:0296` | 126 → `2B87h` | damage flags 含 Electricity 時呼叫 `Protected(0)` | `exact` |

`84h` 不在 overlay 12 `INITEFFPROX` 初始化；Borland module 23 `SPELLS`
於 overlay 22 local `6C74h` 寫入 slot 84。其 handler local `62D7h` 在
`6323h` push `33h` 後透過 spell dispatch 呼叫。這只證明能力及 dispatch；
`ds:A81F < 4` 的完整頻率語意、目標選擇與 terrain line 尚未由 runtime 關閉。

## 工具與驗證契約

- `internal/pc98ovr.Decode` 保存原始 `Code`／`Relocation`，另附加 typed
  `Entries`／`RelocationOffsets`；不覆蓋原 bytes。
- `ResolveStub` 只解析 offset 幾何，呼叫端仍必須證明 pointer segment。
- `cmd/pc98-ovr-audit -resolve-stub 12:0214` 可重現輸出 entry 100、local
  `2404h`；auditor 同時列出每段 fixup 數。
- malformed `CD 3F`、越界 handler、奇數／越界／非遞增 fixup 必須失敗即
  關閉（fail-closed）。
- `scripts/ida/pc98_monster_affect_loader_audit.idc` 保存 overlay 12／22 的
  可重現報告範圍；報告仍須與 raw bytes 交叉驗證。

## 尚未完成

- `4Fh／70h／87h` 尚未接入 remake 所有 pre-damage boundary。
- 怪物 Lightning Bolt 的施放頻率、AI、目標、牆面反彈、逐格動畫與聲音。
- 三神器對提朗瑟克斯效果的 writer／consumer 與終戰 DOS runtime trace。
- HIGH PRIEST `09h／0Ah`、MARGOYLE `77h` 與其餘效果表。
