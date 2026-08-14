# 936 — 等按鍵：`STING` 命令列密碼開啟的開發者後門

- 證據等級：`exact`（逐條讀完；字串與集合常數取自 `START.EXE.asm` 的資料定義）
- 位置：DOS `START.EXE` 的 `16FADh`
- 作法見 spec 783

## `16FADh`（116 條）：等一個按鍵 —— `retf`，回傳 `al`

```pascal
if byte_2117Ch <> 0 then begin              { 非阻塞模式 }
    if KeyPressed then 鍵 := ReadKey else 鍵 := 0;
end else
    鍵 := ReadKey;                          { 阻塞 }

{ ── Ctrl-S：音效開關（永遠有效）── }
if 鍵 = 13h then begin
    if byte_21DAAh <> 2 then begin
        byte_21DABh := byte_21DAAh;         { 記住目前裝置 }
        <sub_17150h>(word_1E76Ch);
        byte_21DAAh := 2;                   { 2 ＝ 關閉 }
    end
    else if byte_21DABh in [0, 1] then begin
        byte_21DAAh := byte_21DABh;         { 還原 }
        <sub_17150h>(word_1E76Eh);
    end;
end;

{ ── 以下三個鍵只有帶 STING 參數才作用 ── }
if ParamStr(1) = 'STING' then begin
    case 鍵 of
      04h:  begin                            { Ctrl-D }
                byte_21181h := not byte_21181h;
                if byte_21181h <> 0 then begin
                    Assign(除錯檔, 'debug.txt');   Rewrite(除錯檔);
                end else
                    Close(除錯檔);
            end;
      1Ah:  if byte_21181h <> 0 then begin    { Ctrl-Z }
                p := 遠指標(dword_226CAh);
                while p <> NIL do begin
                    <sub_16F03h>(p);          { spec 932 的效果鏈傾印 }
                    p := p^[189h];            { 隊伍鏈 }
                end;
            end;
      03h:  <sub_16EA0h>();                   { Ctrl-C }
    end;
end;

{ 把緩衝區裡剩下的鍵吃掉，只保留最後一個 }
if 鍵 <> 0 then
    while KeyPressed do 鍵 := ReadKey;
結果 := 鍵;
```

## ★ `STING` 是出貨版裡的除錯密碼

```
aSting_0   db 5,'STING'
aDebugTxt  db 9,'debug.txt'
```

`ParamStr(1)` 是**命令列的第一個參數**。帶上 `STING` 執行遊戲，
就會啟用三個原本不存在的按鍵。這個字串在 `START.EXE` 裡出現兩次
（另一處在 `seg000:016Eh` 的啟動碼），所以不只這一支在檢查。

| 鍵 | 行為 |
|---|---|
| **Ctrl-D** | 切換除錯記錄：開啟時 `Assign` ＋ `Rewrite` 建立 **`debug.txt`**，關閉時 `Close` |
| **Ctrl-Z** | 除錯記錄開著時，**把全隊每個角色的效果鏈傾印進 `debug.txt`** |
| **Ctrl-C** | 呼叫 `sub_16EA0h`（本規格未讀） |

Ctrl-Z 走 `dword_226CAh` 起的隊伍鏈、next 在 **`+189h`**——
與 spec 913／917 走隊伍時用的同一格吻合。
每個角色交給 spec 932 的 `sub_16F03h`，印出
`who: <名字>  sp#: <法術編號>` 一行一個效果。

**這是一條完整的、可實際使用的除錯路徑**：命令列加 `STING`、
遊戲中按 Ctrl-D 開檔、按 Ctrl-Z 傾印。對 remake 的驗證很有用——
可以拿原版跑出 `debug.txt` 當 oracle 比對效果鏈。

## ★ Ctrl-S 的三個值對上 spec 929 的兩種硬體

`byte_21DAAh` 的合法值由那個 32 bytes 的集合常數界定：

```
byte_16F7D  db 3, 1Fh dup(0)
```

第一個 byte 是 `3` ＝ `0000 0011`，其餘 31 bytes 全 0，
所以集合是 **`[0, 1]`**。加上程式用的 `2`，三個值是：

| 值 | 意義 |
|---|---|
| 0 / 1 | **兩種音效裝置**——正好對上 spec 929 的 PC speaker（`1941Bh`）與 Tandy（`19490h`） |
| 2 | 關閉 |

Ctrl-S 在「目前裝置」與「關閉」之間來回：關閉時把裝置編號存進
`byte_21DABh`，還原時先確認它落在 `[0, 1]` 才寫回——
**所以按 Ctrl-S 不會把音效切成第三種裝置，只會靜音／解除靜音**。

## 吃掉多餘按鍵

結尾那個 `while KeyPressed do 鍵 := ReadKey` 會把鍵盤緩衝區清空，
**只留最後一個**。所以連按會被吃掉，只有最後一次生效。
`鍵 = 0`（非阻塞模式沒按到）時跳過這一段。

## 明確不宣稱

- 沒有宣稱 `<sub_16EA0h>`（Ctrl-C）做什麼。
- 沒有宣稱 `<sub_17150h>` 與 `word_1E76Ch`／`word_1E76Eh` 的內容
  （形狀上是「停止音效」與「恢復音效」兩個曲目編號或指令）。
- 沒有宣稱 `byte_2117Ch`（阻塞／非阻塞）由誰設定。
- 沒有宣稱 `seg000:016Eh` 那一處 `STING` 檢查做什麼。
- 沒有宣稱擴充鍵（掃描碼 0 之後的第二個 byte）怎麼處理——本支只讀一次
  `ReadKey`，沒有處理 `#0` 前綴。
