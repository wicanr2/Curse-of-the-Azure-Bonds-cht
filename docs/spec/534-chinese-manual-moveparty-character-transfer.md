# 第五百三十四輪：中文說明書確認 `MOVEPARTY` 是跨遊戲角色轉移邊界

狀態：`READY`（中文說明書明確證實角色轉移；尚未把 PC-98 raw helper 與該工具
做成同版本 runtime 的一對一閉合）

日期：2026-08-10

## 結論先行

使用者提供的中文《青色枷的詛咒》說明書確實記載「傳送人物」功能，而且畫面
原文是：

```text
Move characters where?
```

它不是把整個劇情存檔直接搬到下一款遊戲，而是把角色資料在 SSI Gold Box
遊戲之間轉移，再於目的遊戲的 `ADD CHARACTER`／加入人物流程中加入角色。
因此先前把 PC-98 `MOVEPARTY` 當作「秘密門」或地圖特殊移動的工作標籤已被本
規格勘誤；`MOVEPARTY` 應優先視為「跨遊戲角色／隊伍資料轉移工具」候選。

這項手冊證據不會自動證明 PC-98 overlay-14 的每一個 B／P／K helper 都是同一
個版本的轉移分支，也不會證明它和騎士事件後 `(13,10)`→`(8,15)` 的地圖路徑
有關。後者仍是獨立的外部地圖／正常移動缺口。

## 原始輸入與頁面證據

| 證據 | SHA-256／工具 | 位址／頁面基準 |
|---|---|---|
| `珍020-青色枷的詛咒.rar` | `63b75989f3e2472fddd7e2e89676580f99a6125470a1501049ca0f09c92eeb5c` | 使用者提供的原始 RAR；唯讀 |
| `011.jpg` | `93707c1084bd96d69b488ebad06b7a00b09c045e02432e76691a8108e1063a` | 掃描印刷頁 3–4 |
| `012.jpg` | `2b99dbeaa123fc2f0bf50077af7281363dae33c8637bc746b28a74d8dc005409` | 掃描印刷頁 5–6 |
| `015.jpg` | `00de1622825089b32517aac0bfd16fea57f7f9fe7f70617cfdbdba048b0d0003` | 掃描印刷頁 11–12 |
| RAR 解壓 | `unar 1.10.1`，Docker 一次性容器 | Big5／中文檔名；未寫入 repo |

### 印刷頁 3–4：傳送人物

中文說明書在印刷頁 3 的「四、傳送人物」前置資料處理章節，說明把人物資料
在磁片／硬碟間準備好；印刷頁 4 的程式畫面直接列出：

```text
Move characters where?
Pool of Radiance to Hillsfar
Pool of Radiance to Curse of the Azure Bonds
Curse of the Azure Bonds to Hillsfar
Hillsfar to Curse of the Azure Bonds
Quit
```

說明書同頁另註明：傳送前必須先把原磁片內的隊伍解散；在《青色枷的詛咒》
中，Paladin／遊俠與 Ranger／流浪者不能傳送到 Hillsfar 與 Pool of Radiance。
這證明的是「角色資料跨遊戲轉移」與限制，不是地圖上的門。

### 印刷頁 11–12：目的遊戲加入外部角色

印刷頁 12 的 `ADD CHARACTER` 說明列出：

```text
FROM WHERE:
CURSE  POOL  HILLSFAR  EXIT
```

其中 `CURSE` 是本作建立的角色，`POOL` 是來自《光芒之池》，`HILLSFAR` 是
來自《幽城寶藏》；若是其他系列遊戲的角色，必須先使用「傳送人物程式」把
資料傳過來。這和印刷頁 3–4 的跨遊戲方向表互相閉合。

## 證據分級與勘誤

- `exact`：中文說明書明文列出四個跨遊戲方向、`Move characters where?`、
  `ADD CHARACTER` 的來源選單，以及來源隊伍解散與職業限制。
- `strong inference`：PC-98 Borland symbol `MOVEPARTY=00C9:0BCCh` 很可能是
  這個「傳送人物」工具或其核心資料流程；名稱、流程目的與手冊語意一致，且
  `MOVEPARTY` 的現有 raw flow 不是普通 `Move` 方向輸入。但仍需同一 PC-98
  版本的 entry selector、畫面字串／磁片 I/O 或 runtime trace 完成一對一閉合。
- `unknown`：B／P／K helper 的正式用途、`DS:7F27h` mode、`+592h`／`+5AAh`
  欄位、四個 `0x014C` writer 是否只是轉移資料檔內的工作欄位，以及它們和
  地圖第三平面的關係。

## 對 remake 與 engine 的影響

1. 不新增 `secret_door`、地圖 `search` 或 `(13,10)`→`(8,15)` 的 movement
   JSON；不能再用 `MOVEPARTY` helper 直接實作地圖通路。
2. 後續應把跨遊戲功能建模成作品中立的 `character_transfer`／角色資料匯入
   邊界：來源遊戲、目的遊戲、角色 record、職業／等級限制與來源資料版本應
   由 game-pack／JSON 宣告，不能把四個方向硬編碼到 CoAB 的 Go State。
3. 仍需解析 DOS／PC-98 的角色資料檔格式、來源／目的磁片 selector、角色
   record round-trip 與失敗條件；在這些證據完成前，先只保留本規格與手冊索引。
4. P0-2 地圖缺口改回追查真正的 external map／wall interaction consumer；
   `MOVEPARTY` 可作角色轉移研究入口，但不再是秘密門的先驗答案。

## 來源限制

這是使用者提供的軟體世界中文說明書掃描檔，頁面文字可直接讀取；掃描頁沒有
在 repo 內重複散布，僅保存原始 RAR 雜湊、頁面檔名與本規格。手冊證明產品
功能與操作流程，不單獨證明特定 PC-98 overlay 的位址語意；binary 結論仍需
遵守 `AGENTS.md` 的原始位址／bytes／consumer／runtime 分級契約。
