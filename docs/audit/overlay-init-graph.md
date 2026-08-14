# overlay unit 初始化鏈（模組相依圖）

由 `scripts/overlay_init_graph.py` 從**原始 bytes** 解出。每個 overlay 的
`0000h` 是它的 unit 初始化段，裡面依序呼叫每個相依 unit 的初始化段，因此
這張表就是模組層級的 uses 關係。

## dos

形狀符合 32 個模組；不符 4 個（overlay-00、overlay-01、overlay-13、overlay-18）。

| 模組 | 直接相依 |
|---|---|
| `overlay-00` | *（形狀不符）* |
| `overlay-01` | *（尾巴不符）* |
| `overlay-02` | `overlay-19`、`overlay-29`、`overlay-33`、`overlay-20`、`overlay-26`、`overlay-16`、`overlay-23`、`overlay-14`、`overlay-24`、`overlay-28`、`overlay-07`、`overlay-21`、`overlay-34`、`overlay-25`、`overlay-03` |
| `overlay-03` | `overlay-34` |
| `overlay-04` | `overlay-21`、`overlay-19`、`overlay-26`、`overlay-24`、`overlay-07`、`overlay-34`、`overlay-22`、`overlay-23`、`overlay-25` |
| `overlay-05` | `overlay-19`、`overlay-07`、`overlay-21`、`overlay-26`、`overlay-06`、`overlay-24`、`overlay-34`、`overlay-22`、`overlay-23`、`overlay-33` |
| `overlay-06` | `overlay-21`、`overlay-19`、`overlay-26`、`overlay-24`、`overlay-07`、`overlay-34` |
| `overlay-07` | `overlay-16`、`overlay-29`、`overlay-23`、`overlay-28`、`overlay-19`、`overlay-24`、`overlay-34`、`overlay-26`、`overlay-33` |
| `overlay-08` | `overlay-19`、`overlay-26`、`overlay-31`、`overlay-32`、`overlay-33`、`overlay-24`、`overlay-22`、`overlay-23`、`overlay-20`、`overlay-34`、`overlay-10`、`overlay-13`+1D9Dh ⚠、`overlay-09`、`overlay-25` |
| `overlay-09` | `overlay-19`、`overlay-31`、`overlay-32`、`overlay-30`、`overlay-26`、`overlay-24`、`overlay-22`、`overlay-23`、`overlay-13`+1D9Dh ⚠、`overlay-34`、`overlay-25` |
| `overlay-10` | `overlay-19`、`overlay-31`、`overlay-30`、`overlay-32`、`overlay-16`、`overlay-26`、`overlay-24`、`overlay-23`、`overlay-13`+1D9Dh ⚠、`overlay-33`、`overlay-17`、`overlay-34`、`overlay-29` |
| `overlay-11` | `overlay-14`、`overlay-22`、`overlay-35`、`overlay-33`、`overlay-29`、`overlay-12`、`overlay-26`、`overlay-34` |
| `overlay-12` | `overlay-24`、`overlay-25`、`overlay-34`、`overlay-23` |
| `overlay-13` | *（尾巴不符）* |
| `overlay-14` | `overlay-23`、`overlay-24`、`overlay-19`、`overlay-28`、`overlay-16`、`overlay-33`、`overlay-29`、`overlay-20`、`overlay-34`、`overlay-26`、`overlay-25` |
| `overlay-15` | `overlay-34`、`overlay-22`、`overlay-20`、`overlay-24`、`overlay-16`、`overlay-26`、`overlay-19`、`overlay-28`、`overlay-23` |
| `overlay-16` | `overlay-33`、`overlay-34`、`overlay-25`、`overlay-23`、`overlay-24` |
| `overlay-17` | `overlay-24`、`overlay-19`、`overlay-22`、`overlay-23`、`overlay-29`、`overlay-33`、`overlay-16`、`overlay-26`、`overlay-32`、`overlay-34`、`overlay-25` |
| `overlay-18` | *（尾巴不符）* |
| `overlay-19` | `overlay-26`、`overlay-29`、`overlay-22`、`overlay-23`、`overlay-21`、`overlay-20`、`overlay-28`、`overlay-24`、`overlay-34`、`overlay-25` |
| `overlay-20` | `overlay-34`、`overlay-22`、`overlay-23`、`overlay-24`、`overlay-26` |
| `overlay-21` | `overlay-26`、`overlay-24`、`overlay-23`、`overlay-34` |
| `overlay-22` | `overlay-24`、`overlay-25`、`overlay-23`、`overlay-26`、`overlay-34` |
| `overlay-23` | `overlay-24`、`overlay-34`、`overlay-25` |
| `overlay-24` | `overlay-33`、`overlay-34`、`overlay-29`、`overlay-28`、`overlay-26`、`overlay-25` |
| `overlay-25` | `overlay-34`、`overlay-26` |
| `overlay-26` | `overlay-29`、`overlay-34` |
| `overlay-27` | *（無）* |
| `overlay-28` | `overlay-33`、`overlay-34`、`overlay-29` |
| `overlay-29` | `overlay-34` |
| `overlay-30` | `overlay-35`、`overlay-34` |
| `overlay-31` | *（無）* |
| `overlay-32` | `overlay-33`、`overlay-34` |
| `overlay-33` | `overlay-34` |
| `overlay-34` | *（無）* |
| `overlay-35` | *（無）* |

## pc98

形狀符合 32 個模組；不符 4 個（overlay-00、overlay-01、overlay-13、overlay-18）。

| 模組 | 直接相依 |
|---|---|
| `overlay-00` | *（形狀不符）* |
| `overlay-01` | *（尾巴不符）* |
| `overlay-02` | `overlay-19`、`overlay-29`、`overlay-33`、`overlay-20`、`overlay-26`、`overlay-16`、`overlay-23`、`overlay-14`、`overlay-24`、`overlay-28`、`overlay-07`、`overlay-21`、`overlay-34`、`overlay-25`、`overlay-03` |
| `overlay-03` | `overlay-34` |
| `overlay-04` | `overlay-21`、`overlay-19`、`overlay-26`、`overlay-24`、`overlay-07`、`overlay-34`、`overlay-22`、`overlay-23`、`overlay-25` |
| `overlay-05` | `overlay-19`、`overlay-07`、`overlay-21`、`overlay-26`、`overlay-06`、`overlay-24`、`overlay-34`、`overlay-22`、`overlay-23`、`overlay-33` |
| `overlay-06` | `overlay-21`、`overlay-19`、`overlay-26`、`overlay-24`、`overlay-07`、`overlay-34` |
| `overlay-07` | `overlay-16`、`overlay-29`、`overlay-23`、`overlay-28`、`overlay-19`、`overlay-24`、`overlay-34`、`overlay-26`、`overlay-33` |
| `overlay-08` | `overlay-19`、`overlay-26`、`overlay-31`、`overlay-32`、`overlay-33`、`overlay-24`、`overlay-22`、`overlay-23`、`overlay-20`、`overlay-34`、`overlay-10`、`overlay-13`+1DD7h ⚠、`overlay-09`、`overlay-25` |
| `overlay-09` | `overlay-19`、`overlay-31`、`overlay-32`、`overlay-30`、`overlay-26`、`overlay-24`、`overlay-22`、`overlay-23`、`overlay-13`+1DD7h ⚠、`overlay-34`、`overlay-25` |
| `overlay-10` | `overlay-19`、`overlay-31`、`overlay-30`、`overlay-32`、`overlay-16`、`overlay-26`、`overlay-24`、`overlay-23`、`overlay-13`+1DD7h ⚠、`overlay-33`、`overlay-17`、`overlay-34`、`overlay-29` |
| `overlay-11` | `overlay-14`、`overlay-22`、`overlay-35`、`overlay-33`、`overlay-29`、`overlay-12`、`overlay-26`、`overlay-34` |
| `overlay-12` | `overlay-24`、`overlay-25`、`overlay-34`、`overlay-23` |
| `overlay-13` | *（尾巴不符）* |
| `overlay-14` | `overlay-23`、`overlay-24`、`overlay-19`、`overlay-28`、`overlay-16`、`overlay-33`、`overlay-29`、`overlay-20`、`overlay-34`、`overlay-26`、`overlay-25` |
| `overlay-15` | `overlay-34`、`overlay-22`、`overlay-20`、`overlay-24`、`overlay-16`、`overlay-26`、`overlay-19`、`overlay-28`、`overlay-23` |
| `overlay-16` | `overlay-33`、`overlay-34`、`overlay-25`、`overlay-23`、`overlay-24` |
| `overlay-17` | `overlay-24`、`overlay-19`、`overlay-22`、`overlay-23`、`overlay-29`、`overlay-33`、`overlay-16`、`overlay-26`、`overlay-32`、`overlay-34`、`overlay-25` |
| `overlay-18` | *（尾巴不符）* |
| `overlay-19` | `overlay-26`、`overlay-29`、`overlay-22`、`overlay-23`、`overlay-21`、`overlay-20`、`overlay-28`、`overlay-24`、`overlay-34`、`overlay-25` |
| `overlay-20` | `overlay-34`、`overlay-22`、`overlay-23`、`overlay-24`、`overlay-26` |
| `overlay-21` | `overlay-26`、`overlay-24`、`overlay-23`、`overlay-34` |
| `overlay-22` | `overlay-24`、`overlay-25`、`overlay-23`、`overlay-26`、`overlay-34` |
| `overlay-23` | `overlay-24`、`overlay-34`、`overlay-25` |
| `overlay-24` | `overlay-33`、`overlay-34`、`overlay-29`、`overlay-28`、`overlay-26`、`overlay-25` |
| `overlay-25` | `overlay-34`、`overlay-26` |
| `overlay-26` | `overlay-29`、`overlay-34` |
| `overlay-27` | *（無）* |
| `overlay-28` | `overlay-33`、`overlay-34`、`overlay-29` |
| `overlay-29` | `overlay-34` |
| `overlay-30` | `overlay-35`、`overlay-34` |
| `overlay-31` | *（無）* |
| `overlay-32` | `overlay-33`、`overlay-34` |
| `overlay-33` | `overlay-34` |
| `overlay-34` | *（無）* |
| `overlay-35` | *（無）* |
