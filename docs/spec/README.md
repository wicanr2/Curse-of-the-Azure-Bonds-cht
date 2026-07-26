# 格式規格

相關中文資料：[繁中遊玩手冊](../manual/curse-of-the-azure-bonds-zh-TW.md) ・ [中文金盒子歷史筆記](../history.md)

本目錄只收錄由原始映像、執行觀察、反組譯或可重現工具支持的規格。

狀態定義：

- `DRAFT`：有初步觀察，但仍有未驗證的關鍵假設。
- `READY`：可供實作，欄位、邊界、錯誤行為與驗證案例已足夠明確。

目前規格：

- [第一輪素材與格式盤點](./round-01-inventory.md)（`DRAFT`）
- [第二輪 DAX 容器與 ECL 文字取樣](./round-02-dax-container.md)（`DRAFT`）
- [第三輪 DOS loader 與 GAME.OVR](./round-03-loader-and-overlay.md)（`DRAFT`）
- [第四輪 Go 核心層與繁中資源](./round-04-go-core.md)（`READY`：限 DAX container／locale）
- [第五輪 ECL operand framing](./round-05-ecl-operands.md)（`DRAFT`）
- [第六輪 ECL 安全 trace walker](./round-06-ecl-trace.md)（`DRAFT`）
- [第七輪 ECL packed text extraction](./round-07-ecl-text.md)（`DRAFT`）
- [第八輪繁中開場狀態核心](./round-08-localized-state.md)（`DRAFT`）
- [第九輪 Ebiten opening prototype](./round-09-ebiten-opening.md)（`DRAFT`）
- [第十輪 ECL data-driven opening](./round-10-data-driven-opening.md)（`DRAFT`）
- [第十一輪 ECL branch target graph](./round-11-ecl-branch-graph.md)（`DRAFT`）
- [第二十六輪 ECL5 real NEWECL regression](./round-26-ecl5-newecl-regression.md)（`READY`：限已定位 transition entry）
- [第二十七輪 Shadowdale 荒野入口](./round-27-shadowdale-map-entry.md)（`READY`：限入口狀態與輸入 contract）
- [第二十八輪 Shadowdale 場所 menu](./round-28-shadowdale-place-menu.md)（`READY`：限選項與 state-level event contract）
- [第二十九輪 TREASURE bounded prefix](./round-29-treasure-prefix.md)（`READY`：限 operand framing 與安全前綴）
- [第三十輪 COMBAT request signal](./round-30-combat-request.md)（`READY`：限控制轉移 signal）
- [第三十一輪 AD&D combat core](./round-31-combat-core.md)（`READY`：限可注入骰點的核心規則）
- [第三十二輪 ECL monster spawn descriptors](./round-32-ecl-monster-spawn.md)（`READY`：限 ECL command descriptor）
- [第三十三輪 MON*CHA monster record](./round-33-monster-cha-record.md)（`READY`：限固定 record offsets 與 raw combat fields）
- [第三十四輪 MON*ITM／MON*SPC records](./round-34-monster-items-affects.md)（`READY`：限 raw record parser）
- [第三十五輪 monster item／effect 繁中名稱](./round-35-monster-item-localization.md)（`READY`：限本輪實際觀察 IDs）
- [第三十六輪 ECL-to-enemy encounter adapter](./round-36-ecl-encounter-adapter.md)（`READY`：限 enemy fighter 建立）
- [第三十七輪 可操作戰鬥狀態與 Ebiten 畫面](./round-37-playable-combat-state.md)（`READY`：限戰鬥垂直切片）
- [第三十八輪 ECL encounter 到 Battle 的資料橋](./round-38-ecl-encounter-to-battle.md)（`READY`：限 ECL1 direct-entry）
- [第三十九輪 PROGRAM 外部 routine 邊界](./round-39-program-boundary.md)（`READY`：限 VM 控制轉移）
- [第四十輪 遊戲內冒險手札與 CAMP state](./round-40-journal-and-camp-state.md)（`READY`：限資料呈現與控制邊界）
- [第四十一輪 可翻頁的繁中冒險手札](./round-41-journal-pages.md)（`READY`：限八頁摘要與 UI 導航）
- [第四十二輪 party 保存與 CAMP 恢復 state](./round-42-party-camp-state.md)（`READY`：限 party HP boundary）
- [第四十三輪 角色建立規則核心](./round-43-character-creation-rules.md)（`READY`：限 validation）
- [第四十四輪 繁中角色建立 UI](./round-44-character-creation-ui.md)（`READY`：限 starter slice）
- [第四十五輪 繁中角色姓名輸入](./round-45-character-name-input.md)（`READY`：限 Unicode 姓名）
- [第四十六輪 角色建立能力值編輯](./round-46-ability-editor.md)（`READY`：限能力值 slice）
