# ECL `CALL`（opcode `2Dh`）external routine registry

由 `scripts/ecl_call_registry.py` 產生，不要手改。
selector 與原始 ECL operand 的關係是 `operand = (selector + bias) mod 10000h`；
bias 由 handler 內的 `sub ax, NNNNh` 直接讀出，不是假設值。

- **DOS**：handler `2F02h..3073h`，bias `7FFFh`，認得的 selector 7 個，需人工讀 0 個。其餘 selector 走到 epilogue（不做事就返回）。
- **PC98**：handler `30C6h..3237h`，bias `7FFFh`，認得的 selector 7 個，需人工讀 0 個。其餘 selector 走到 epilogue（不做事就返回）。

| ECL operand | selector | DOS 分支 | PC98 分支 |
|---|---|---|---|
| `8000h` | `0001h` | `2FAFh` | `3173h` |
| `8001h` | `0002h` | `2FBFh` | `3183h` |
| `B200h` | `3201h` | `2FCFh` | `3193h` |
| `C018h` | `4019h` | `300Eh` | `31D2h` |
| `C01Eh` | `401Fh` | `3002h` | `31C6h` |
| `2E10h` | `AE11h` | `2F31h` | `30F5h` |
| `6803h` | `E804h` | `3035h` | `31F9h` |

`分支` 是該 selector 在各自 overlay-02 內的 code-local 位址；兩平台位址
不同是正常的（見 spec 560），對應依據是 selector 值本身。
