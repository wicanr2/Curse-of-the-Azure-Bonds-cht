# 第二百五十輪：半獸人 race／DOS parser

## 狀態

READY

## 證據

reference `Gbl.RaceClasses` 明確列出半獸人的 cleric、fighter、thief 與 multi-class
組合；重新依 `ClassId` index 核對後，`race_ages` 的 half-orc row 提供 cleric
`20 + 1d4`、fighter `13 + 1d4`、thief `20 + 2d4`，paladin／ranger／magic-user entry
為零。外部 AD&D 1e age table 也給出同一 fighter age。`Limits`
也提供 half-orc age brackets 與共用 age effects。

## 實作邊界

- `party.RaceHalfOrc` 與 raw DOS race `6` 已可解析、驗證、序列化並使用正常 combat icon size。
- 建立選單接入半獸人 cleric／fighter／thief；fighter 的 `13+1d4` 已修正為
  reference `ClassId=2`，不再誤套到本專案的 ranger index。
- half-orc multi-class、alignment／level limits 與完整原版建立流程仍待接續。

## 驗證

- `TestHalfOrcSingleClassEvidence` 驗證 `20+1d4`／`13+1d4`／`20+2d4` age table 與 class restriction。
- `TestParseDOSHalfOrcRace` 驗證 raw race `6` 進入 DOS player／Character。
