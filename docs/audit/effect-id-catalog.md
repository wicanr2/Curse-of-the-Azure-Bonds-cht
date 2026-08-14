# 效果編號 ↔ 訊息 對照（線索表）

由 `scripts/effect_id_catalog.py` 產生，範圍 `dos overlay-22`。
每一列是一支函式：它對哪些效果編號做了事，以及它裡面有哪些訊息字串。

**同一支有多個編號或多句訊息時本表不強配**——這是線索不是結論，
引用某個編號的語意之前要回去讀那支函式。

`23#3` ＝ `overlay-23` entry#3（解除）、`23#16` ＝ entry#16（查詢）、
`24#27` ＝ `overlay-24` entry#27（找節點）。

| 函式 | 編號 | 訊息 |
|---|---|---|
| `08A1h` | `10h`(24#27) | — |
| `1263h` | `4Ah`(24#27) | can't be cast here... / Lose it?  / That Item / is a combat-only item... / Use it?  / miscasts / casts / Abort Spell?  / Spell Aborted |
| `1EC9h` | `0Bh`(24#27) | is unaffected / is charmed |
| `20E1h` | `0Ch`(24#27) `0Ch`(23#3) | has been reduced |
| `22B4h` | `35h`(24#27) | falls asleep |
| `24CDh` | `37h`(24#27) | is affected |
| `2F5Eh` | `21h`(23#16) | can see |
| `2FD1h` | `22h`(23#16) `2Bh`(23#16) `2Ch`(23#3) `1Fh`(23#3) `32h`(23#16) `39h`(23#3) | — |
| `317Eh` | `5Bh`(23#3) `28h`(23#3) | is affected |
| `3599h` | `24h`(23#16) | is un-cursed / has an item un-cursed |
| `3EE2h` | `2Ah`(23#16) | is Speedy |
| `4194h` | `37h`(24#27) `37h`(23#3) `16h`(23#3) `0Fh`(23#3) | is unpoisoned / is unaffected |
| `4311h` | `03h`(24#27) | smashes them flat |
| `44D3h` | `20h`(23#3) `37h`(23#3) | is raised |
| `4791h` | `0Bh`(24#27) | is charmed |
| `48E4h` | `3Ah`(24#27) `90h`(24#27) `8Bh`(24#27) `90h`(23#3) `8Bh`(23#3) | teleports |
| `4DA0h` | `1Bh`(24#27) `2Ah`(24#27) | is clumsy / is slowed |
| `5590h` | `44h`(24#27) | — |
| `57E3h` | `58h`(23#3) | Breathes! |

## 各編號出現次數

`03h`×1 `0Bh`×2 `0Ch`×2 `0Fh`×1 `10h`×1 `16h`×1 `1Bh`×1 `1Fh`×1 `20h`×1 `21h`×1 `22h`×1 `24h`×1 `28h`×1 `2Ah`×2 `2Bh`×1 `2Ch`×1 `32h`×1 `35h`×1 `37h`×4 `39h`×1 `3Ah`×1 `44h`×1 `4Ah`×1 `58h`×1 `5Bh`×1 `8Bh`×2 `90h`×2
