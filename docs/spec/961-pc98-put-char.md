# 961 — PC-98 的單字元輸出：全形不跨行，退格會跨到上一行行尾

- 證據等級：`exact`（逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `19493h`
- 作法見 spec 783

## `19493h`（157 行）

`al` ＝ 字元。`spec 952`／`spec 953` 那條路是**批次**版；本支是**逐字元**版，
spec 945／948／951 都呼叫它。

```pascal
es := word_280D4h;  di := word_280D8h;  bx := word_280D6h;  cl := byte_280C6h;
<sub_195FCh>();
ah := byte_280E3h;                                { 前導旗標 }

if ah <> 0 then begin                             { ★ 這次是全形的後續 }
    byte_280E3h := 0;
    <sub_18FA3h>(屬性);
    es:[bx＋di] := 屬性;   es:[bx＋di＋2] := 屬性;
    <sub_19791h>();  xchg(ah, al);  al := al − 20h;
    stosw;                                        { 左半格 }
    ax := ax or 8000h;  stosw;                    { 右半格 }
    dl := dl ＋ 2;
    if dl > 右界 then begin dl := 左界;  換行 end;
end
else if (byte_280D0h <> 0) and <sub_1977Eh>(al) then begin
    byte_280E3h := al;                            { 記下前導 }
    if dl >= 右界 then begin                       { ★ 放不下整個全形字 }
        <sub_19BBCh>();  換行();  dl := 左界;  <sub_19604h>();
    end;
end
else case al of
    07h: <sub_19A07h>();                          { BEL }
    08h: 退格（見下）;
    0Dh: if dl <> 左界 then dl := 左界;
    0Ah: 換行();
  else if al >= 20h then begin
        <sub_18FA3h>(屬性);   es:[bx＋di] := 屬性;
        stosw;                                    { 字碼 }
        if byte_280DAh <> 50h then stosw;         { 40 欄模式再一次 }
        inc(dl);
        if dl > 右界 then begin dl := 左界;  換行 end;
    end;
end;
```

## ★ 退格會跨到上一行的行尾

```pascal
{ al = 08h }
if dl <> 左界 then dec(dl)                        { 一般情況：往左退一格 }
else if dh <> 上界 then begin                     { ★ 已在行首 }
    dec(dh);                                      { 退到上一行 }
    dl := 右界;                                   { 跳到行尾 }
    if cs:byte_19D01h <> 0 then begin             { ★ 全形模式 }
        讀出 (dh, 右界) 那一格;
        if 那一格 <> 20h then dec(dl);            { 不是空白 → 再退一格 }
    end;
end;
```

`cs:byte_19D01h` 就是 **spec 945 的退格常式設進去的那個 self-modifying 旗標**
——這裡是它唯一的讀取端。旗標立著時，退到上一行行尾之後**還要看那一格
是不是空白**：不是空白就再退一格，因為那裡是一個全形字的右半。

spec 945 的待答項（「`cs:byte_19D01h` 由誰讀取」）到此關閉。

## 全形不跨行

與 spec 952（批次版）相同的規則：**遇到前導位元組時先檢查 `dl >= 右界`**，
放不下就先換行，讓整個全形字落在下一行。
兩支各自實作了一份，**行為一致但程式碼是獨立的**。

## 屬性寫兩格

全形字的屬性平面**兩格都寫**（`es:[bx＋di]` 與 `es:[bx＋di＋2]`），
與 spec 953 的批次版一致。

## 明確不宣稱

- 沒有宣稱 `<sub_195FCh>`／`<sub_19604h>`／`<sub_19BBCh>`／`<sub_195B6h>`／
  `<sub_19A07h>`／`<sub_19B7Dh>`／`<sub_19B9Ch>`／`<sub_199F5h>`／
  `<sub_18FA3h>`／`<sub_19791h>` 的內部。
- 沒有宣稱 40 欄模式下全形字為什麼不像單字元那樣重複 `stosw`
  （本支只在單字元分支有那一行）。
