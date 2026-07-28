# 第三百二十三輪：DOS 戰鬥畫面量測校正

狀態：`READY`

## Oracle 與量測

以 DOS 版 `combat-aim`、`fight-black-dragon` 截圖為 visual oracle。來源 PNG
排除視窗標題後是 640×400，即原生 320×200 的 2× nearest-neighbour：

- <https://simeonpilgrim.com/blog/static/977b71ee/combat-aim%402x.png>
- <https://simeonpilgrim.com/blog/static/8d8a2230/fight-black-dragon%402x.png>

原生 320×200 的上半部固定為：

| 區域 | 原生矩形 | 640×480 renderer 矩形 |
|---|---:|---:|
| 戰場 | `(8,8,168,168)` | `(16,16,336,336)` |
| 中央石框 | `x=176..183` | `x=352..367` |
| 右側資訊 | `(184,8,128,168)` | `(368,16,256,336)` |
| 下側石框 | `y=176..183` | `y=352..367` |
| 中文戰鬥紀錄 | 原版無 | `(0,368,640,80)` |
| 原版式 footer | `(0,184,320,16)` | `(0,448,640,32)` |

戰場正好是 7×7 個隱藏的 24×24 movement cells；2× 後每格 48×48。原版不畫
checkerboard 或 debug grid，而是整片黑／灰地板加上牆、階梯與轉角 terrain。

## Renderer contract

- 上方 368px 必須可縮小 50% 後套疊原版 320×184 幾何，不為中文任意拉高。
- 中文新增的 80px 只作戰鬥紀錄；最後 32px 保留兩列 16px footer。
- 右欄只常駐 active combatant：青色姓名、綠色 label、黃色數值。target 移到 footer。
- 一般作用中角色使用單一 48×48 白框；不畫紅／藍 team bars。
- 大型敵人的原版 occupancy／anchor 未解出前，不畫錯誤的一格 target box。
- 框架使用 16px 石框比例與硬邊 EGA palette；目前程序式裂紋只是暫代，
  原始 frame tiles 仍需從資源／reference routine 找出。
- combat terrain selector、斜牆拼接、樓梯及大型怪物 footprint 是下一個 P1
  reverse-engineering boundary，不能宣稱已完成。

## 驗收

- `docs/screenshots/gold-box-layout-combat.png` 無可見棋盤格、無右欄 target card。
- 戰場起點 `(16,16)`、尺寸 336×336；divider 16px；右欄 256px。
- 完整 `go test ./...` 在 Docker／Xvfb 通過。

