# 第一百零二輪：DOS player sidecar bundle importer

狀態：`READY`（限 `.SAV/.GUY` + optional `.FX/.SWG` character bundle）

公開資料確認原版角色由一個必要的 `.SAV`（party save）或 `.GUY`（party 外角色）record，以及可選的 `.FX` effects、`.SWG` inventory sidecars 組成。`party.DOSPlayerFiles`／`ParseDOSPlayerFiles` 將這三個已分開取得的 byte streams 組成一個 `Character`：

- required `Record` → `ParseDOSPlayerRecord`。
- optional `Effects` → `ApplyEffects`／`Character.Effects`。
- optional `Inventory` → `ApplyInventory`／`Character.Equipment`。
- gold、gems、jewelry 也保存到 `Character`。

本輪不解析 `SAVGAM?.DAT` 的 party／area container，也不從文件名推導 save slot 或 pointer address space。那一層必須先取得實際 sample bytes 與 header/record boundary 證據，再接到目前的 versioned remake save。

格式參考：[CoAB save character file description](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)。
