# 戰役測試走的每一步，玩家按得出來嗎

由 `cmd/player-path-audit` 產生，不要手改。方法與判讀見 spec 1189。

★ 補的是「開場到結局」剩下的那半個缺口：spec 1188 已經讓真的前端把戰役的每個檢查點**畫**了一張，但**輸入那一層**還沒有證明——戰役直接呼叫 `state.X()`，玩家按的是鍵盤。前端的 `Update()` 到不了的方法，就沒有任何按鍵組合到得了。

⚠ **方向不對稱，不要反著讀。** 這是靜態可達性：證明得了「前端**沒有**那條路」（硬缺口），證明不了「有那條路所以玩得到」——那一行可能藏在永遠不成立的條件底下。

| 指標 | 數字 |
|---|---:|
| 戰役用到的 `State` 進入點 | 57 |
| 前端直接呼叫的 `State` 方法（輸入側）| 148 |
| 再沿 `State` 內部呼叫展開後 | 478 |
| `Draw()`（顯示）可達閉包裡的 | 129 |
| `main()` 啟動流程裡的 | 383 |
| 其中直接寫在 `Update()` 本體的 | 181 |
| 自動解出的單行別名 | 27 |
| 判定為**會改狀態**的 `State` 方法 | 344 |
| **前端沒有路可以到的** | 1 |

| `State` 進入點 | 戰役用幾處 | 玩家到得了 | 說明 |
|---|---:|---|---|
| `AddCreationCharacter` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `AdvanceGameTimeHours` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `Apply` | 2 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `BashDungeonDoor` | 5 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `CanMoveDungeon` | 6 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `CloseJournal` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `CombatAct` | 40 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `CombatActive` | 34 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `CombatTargets` | 3 | ✅ 按得出來 | 別名 → `livingBySide`（單行轉呼叫），前端走那一支 |
| `ConsumeECLCallRequests` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `ConsumeMusicEvents` | 7 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `Continue` | 98 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `DungeonDoorMenuOptions` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `DungeonGeometryView` | 9 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `EnterDungeonCamp` | 2 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `FinishCharacterCreation` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `LookDungeonLocation` | 2 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `MoveDungeon` | 31 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `OpenJournal` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `RunDungeonExitLifecycle` | 8 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `RunDungeonLifecycle` | 50 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `SavePartyFile` | 2 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `SearchDungeonLocation` | 7 | ✅ 按得出來 | 別名 → `LookDungeonLocation`（單行轉呼叫），前端走那一支 |
| `Select` | 292 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `SetDungeonGeometryView` | 2 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `SetParty` | 4 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `StartCombat` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `StartEncounter` | 1 | ✅ 按得出來 | 別名 → `StartEncounterWithAffects`（單行轉呼叫），前端走那一支 |
| `ToggleDungeonSearch` | 2 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `TurnDungeonWithGrid` | 33 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `applyECLCallSignals` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `applyECLNPCSignals` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `applyECLTreasureSignals` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `dungeonExternalExit` | 2 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `dungeonSearchEdgeIDs` | 3 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `searchEdgeDiscovered` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `syncDungeonECLRegisters` | 1 | ✅ 按得出來 | `Update()` 的可達閉包裡有 |
| `DungeonSearchActive` | 2 | — 讀取器差異 | **不改狀態**：前端用別的方式讀同一件事（讀欄位、或它已經被投影進 `Choices`） |
| `MoneyPool` | 6 | — 讀取器差異 | **不改狀態**：前端用別的方式讀同一件事（讀欄位、或它已經被投影進 `Choices`） |
| `PendingTreasureItems` | 15 | — 讀取器差異 | **不改狀態**：前端用別的方式讀同一件事（讀欄位、或它已經被投影進 `Choices`） |
| `ShopOffers` | 5 | — 讀取器差異 | **不改狀態**：前端用別的方式讀同一件事（讀欄位、或它已經被投影進 `Choices`） |
| `TreasurePool` | 2 | — 讀取器差異 | **不改狀態**：前端用別的方式讀同一件事（讀欄位、或它已經被投影進 `Choices`） |
| `CombatFighters` | 35 | — 觀測點 | 前端在 `Draw()` 讀它 ⇒ 這是**看**的東西，不是按的東西 |
| `GameTimeDisplay` | 1 | — 觀測點 | 前端在 `Draw()` 讀它 ⇒ 這是**看**的東西，不是按的東西 |
| `PartyFighters` | 4 | — 觀測點 | 前端在 `Draw()` 讀它 ⇒ 這是**看**的東西，不是按的東西 |
| `CombatStatus` | 17 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `MessageContainsGamePackText` | 7 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `OriginalChoiceIndex` | 7 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `SetECLSeed` | 3 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `SetGeoCatalog` | 1 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `SetMonsterAffectsForECL` | 4 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `SetMonsterItemsForECL` | 4 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `SetMonsterRecords` | 1 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `SetMonsterRecordsForECL` | 10 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `SetTreasureItemBlocks` | 4 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `StartDungeonStoryPreview` | 5 | — 啟動接線 | 前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作 |
| `NextJournalPage` | 1 | **前端沒有這條路** | 會改狀態而前端到不了 ⇒ 玩家按不出來 |

