# Protocol coverage

What this server implements for **Hot Wheels: World's Best Driver** (Wii U),
and how sure we are of each. No packet capture of this title exists yet, so
"status" is about inference confidence from `Game.rpx`'s strings/symbols, not
observation. See [RECON.md](RECON.md).

## Authentication endpoint

| Protocol | Coverage | Status |
|---|---|---|
| Ticket Granting | Login / LoginEx / RequestTicket. `ValidateLoginData` accepts the account-server token (or anything, in local mode). | Stock common handler. |

`PRUDPV1Settings.LegacyConnectionSignature` (`true`) and
`ByteStreamSettings.UseStructureHeader` (`false`) are **unverified
placeholders** - see RECON.md. `nex/authentication.go` dumps the raw
`LoginEx` parameter bytes to stdout for hand-checking against a future
capture, the same way every sibling server in this project does.

## Secure endpoint

| Protocol | Coverage | Status |
|---|---|---|
| Secure Connection | baseline handshake, insecure `Register` | Stock. Required for any secure endpoint. `CreateReportDBRecord` left unset - only used by the rarely-called `SendReport`, not required for the core auth/ranking flow. |
| Utility | baseline (`AcquireNexUniqueID` etc.) | Stock. Not confirmed linked by the RPX strings, registered as cheap insurance. |
| Ranking | `GetRanking` (Range/Own/Friends/Near/FriendNear modes), `UploadScore`, common-data get/upload, plus hand-written `GetApproxOrder`/`GetStats`/`ChangeAttributes`/`DeleteScore` (no common-go implementation exists for these four - confirmed by reading `nex-protocols-common-go` v2.6.1's source). | Postgres-backed (`ranking.*` schema). Friend-scoped modes degrade to the non-friend query (no Friends protocol linked, no friends system). `ChangeAttributes`' exact `Groups`/`ModificationFlag` semantics are inferred from doc comments in `ranking/constants/modification_flag.go`, **not confirmed against a real capture**. |
| Health / Monitoring | `PingDaemon` / `PingDatabase` liveness checks | Stock, real. |
| AccountManagement | registered, no handlers assigned | **Stub - NotImplemented.** No common-go implementation exists; every method already replies `Core::NotImplemented` when its handler is nil. |
| RemoteLogDevice | registered, no handlers assigned | **Stub - NotImplemented.** Same mechanism as above. |
| Subscription | registered, no handlers assigned | **Stub - NotImplemented.** Same mechanism as above. |

## Deliberately not registered yet

| Protocol | Why |
|---|---|
| MatchMaking / MatchMakingExt / MatchmakeExtension / NAT Traversal | `wuNetworkOnline::createGame/getGameList/startMatch` is real, Wii U-specific compiled code, but its reachability in the shipped retail build is unproven (see RECON.md). Stage-5 follow-up once Ranking is console-verified - not wired defensively like Trine 2/Sochi 2014/Rio 2016 because there is no confirmed evidence yet either way, only ambiguous signal. |
| DataStore | No `GhostUpload`/`GhostDownload` symbols anywhere - ghosts appear to be stored locally, not server-synced. No evidence to build against at all, unlike Xenoblade Chronicles X. |
| Messaging / MessageDelivery | No evidence of use in this title. |

## First things to check once a real capture exists

1. Whether `LoginEx` succeeds with the placeholder `LegacyConnectionSignature`/
   `UseStructureHeader` guesses, or needs flipping like Xenoblade Chronicles X
   and Hyrule Warriors both did.
2. The real NEX version (`globals/config.go`'s `3.4.0` is a guess) - affects
   `RankingRankData`/`RankingScoreData` field layout the same way it did for
   Sochi 2014's `Ranking` library-version bump.
3. Whether `wuNetworkOnline`'s matchmaking calls actually reach this server
   during real online play, or whether the game is local-multiplayer-only as
   one (low-reliability) secondary source claims.
4. Whether both region game server IDs are actually needed, or whether (like
   Xenoblade Chronicles X and Art of Balance) one shared ID covers every
   region - see RECON.md.

## Not implemented

- Any matchmaking/online-race session layer - see "Deliberately not
  registered yet" above.
- Ghost/replay upload - no evidence it is server-backed.
- `GetRankingByPIDList` / `GetRankingByUniqueIDList` / `DeleteAllScores` /
  `ChangeAllAttributes` / `DeleteCommonData` - no evidence any of these are
  called by this title; left unregistered (the library answers
  `Core::NotImplemented` on its own for a stray call).
