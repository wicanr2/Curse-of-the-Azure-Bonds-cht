# ECL bank 1 計算值 → record 欄位對照

由 `scripts/ecl_bank1_field_map.py` 產生，不要手改。
base pointer 是 `DS:9594h` 的 far pointer；偏移是該 record 內的位移。
**本表只證明位址對映與存取寬度，不證明欄位語意。**

鏈上共 32 個比較，抽出 22 筆對映。

| ECL 位址 | record 偏移 | 寬度 | 延伸 |
|---|---|---|---|
| `7C15h` | `+013h` | byte | 零 |
| `7C18h` | `+019h` | byte | 零 |
| `7C72h` | `+074h` | byte | 符號 |
| `7C73h` | `+075h` | byte | 符號 |
| `7C9Bh` | `+0E0h` | byte | 零 |
| `7CA0h` | `+0E5h` | byte | 符號 |
| `7CACh` | `+0E9h` | byte | 零 |
| `7CB8h` | `+0F7h` | byte | 零 |
| `7CBBh` | `+0FBh` | word | 零 |
| `7CBDh` | `+0FDh` | word | 零 |
| `7CBFh` | `+0FFh` | word | 零 |
| `7CC1h` | `+101h` | word | 零 |
| `7CC3h` | `+103h` | word | 零 |
| `7CC9h` | `+116h` | byte | 符號 |
| `7CD6h` | `+119h` | byte | 符號 |
| `7CD8h` | `+11Bh` | byte | 符號 |
| `7CE4h` | `+193h` | byte | 零 |
| `7CF7h` | `+13Ch` | word | 零 |
| `7CF9h` | `+13Eh` | byte | 零 |
| `7D1Bh` | `+1A6h` | byte | 零 |
| `7ECFh` | `+01Bh` | byte | 零 |
| `7F3Eh` | `+67Ch` | word | 零 |

⚠ 比較數與抽出數不同（32 vs 22）：有分支不是單純的欄位讀取，
必須逐一人工確認，不得假設漏掉的那幾筆不存在。
