# 原作的換曲點，remake 接上幾個

由 `cmd/music-change-points` 產生，不要手改。換曲點的來源是 `docs/audit/pc98-music-triggers.md`（PC-98 的 Borland 符號表），對照表在 `internal/audiomap`。

| 項目 | 數 |
|---|---:|
| 原作換曲點 | 13 |
| remake 有落點 | 13 |
| 被選到的相異曲目 | 12 |
| game pack 宣告的曲目 | 12 |

⚠ 「有落點」＝ pack 有那首曲子而且有一條指向它的綁定。**不表示**在原作會發的那一刻發——那要實機比對。

| 位置 | 事件 | 曲目 | remake 的 context | 有落點 |
|---|---|---:|---|---|
| `常駐 9486h` | 派曲表：城鎮 | 3 | （照 ECL 段綁） | ✅ |
| `常駐 94AEh` | 派曲表：地城三 | 4 | （照 ECL 段綁） | ✅ |
| `常駐 94C4h` | 派曲表：村莊 | 6 | `pc98-town-services-menu` | ✅ |
| `常駐 94CBh` | 派曲表：荒野 | 5 | （照 ECL 段綁） | ✅ |
| `常駐 94E2h` | 派曲表：散提爾堡城壁 | 8 | （照 ECL 段綁） | ✅ |
| `常駐 94F9h` | 派曲表：盜賊公會 | 9 | （照 ECL 段綁） | ✅ |
| `常駐 9514h` | 派曲表：地城 | 12 | （照 ECL 段綁） | ✅ |
| `overlay-01 093Ch` | 開場 | 1 | `pc98-title` | ✅ |
| `overlay-17 GEN 0B08h` | 角色建立 | 2 | `pc98-character-creation` | ✅ |
| `overlay-10 COMPREP 1DA1h` | 開戰 | 7 | `pc98-combat` | ✅ |
| `overlay-10 COMPREP 1D97h` | 開戰且 LOADMONNUM == 47h | 11 | `pc98-combat-dungeon-two` | ✅ |
| `overlay-05 POSTCOM 1955h` | 全滅 | 2 | `pc98-party-wipe` | ✅ |
| `overlay-18 168Dh` | 結局 | 10 | `pc98-ending` | ✅ |
