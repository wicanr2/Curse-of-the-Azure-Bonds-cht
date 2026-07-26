# 第七十八輪：角色建立玩家 icon 預設

狀態：`READY`（限新建角色 icon default 與 remake save 欄位）

## 已確認

- reference 新建 `Player` 的 `head_icon` 與 `weapon_icon` 保持 0；`icon_id` 初始為 0x0A。
- race switch 設定 `icon_size`：halfling／dwarf／gnome 為 1（small），elf／half-elf／human 為 2（normal）。
- `LoadPlayerCombatIcon` 再以 CHEAD head block 與 CBODY body block 合成 combat icon。

## 本輪成果

- `party.Character.IconSize` 成為 JSON-compatible 欄位；舊 save 沒有欄位時依 race 推導，保留相容性。
- `Character.Fighter()` 現在填入 `HasPartyIcon=true`、head/body block 0 與正確 icon size。
- `State.SetParty` 的 fallback 改為原作 head/body block 0，不再用 party slot 假造不同角色外觀。
- 對六種 race 的 icon size 與 fighter projection 加入 regression tests。

## 邊界

`icon_id` 的 runtime combat slot allocation、角色編輯器的 head／weapon 選擇、recolor palette 與原始 DOS save record 尚未完成；本輪只確定並保存新建角色的初始 icon 欄位。

## 驗證

```sh
go test ./...
go run ./scripts
```
