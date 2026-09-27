# Recon

Evidence trail behind every constant and design choice in this server, so
this session's research survives into the repo.

## The executable

`Game.rpx` (13,599,808 bytes, provided by the operator) is a Cafe OS RPX (ELF32BE,
34 sections, several zlib-compressed per the `SHF_RPL_ZLIB` flag). Decompressed
cleanly to a 54,221,412-byte flat blob with a non-stripped `.strtab`
(~151k symbol-name strings).

**This is genuinely Hot Wheels: World's Best Driver**, confirmed via:

- `.rodata`: `"Hot Wheels"` / `" World's Best Driver"` / `"HotWheels"` /
  `"MattelLogo"` / `"FirebrandLogo"` (x2) - plus leftover NASCAR-engine asset
  names (`nascar_arcade_logo`, `nascar_highlight_animation`) from Firebrand's
  prior engine lineage, reused for this title.
- `.strtab`: dozens of embedded MSVC debug source paths of the form
  `D:\Projects\US\TeamHotWheels\Game\Code\BackEnd\gaAI\gaAI.cpp`,
  `...\gaCamera\gaCameraChase.cpp`, `...\gaCar\gaCarManager.cpp`,
  `...\gaColl\gaCollManager.cpp` - internal project name **"TeamHotWheels"**,
  `ga*` class prefix throughout. Not guessable metadata - this is the
  compiler's own embedded source-path debug info.

A `meta.xml` sitting in the same scratch folder as `Game.rpx` describes
**Art of Balance** (title_id `0005000010149400`, product code `WUP-P-WABP`,
publisher Shin'en) - a stale leftover from that project (finished the same
day this one started). It does not describe `Game.rpx` and was ignored.

## Linked NEX surface

Real, substantial NEX code is compiled in, not a stub:

- `nn::nex::RankingClient` with `Bind`/`Unbind`/`GetRanking`/
  **`GetApproxOrder`**/**`GetStats`**/**`ChangeAttributes`**/**`DeleteScore`**/
  `GetCommonData`. `Bind`/`Unbind` have no corresponding wire method in
  `nex-protocols-go`'s `ranking/protocol.go` - client-local category-binding
  bookkeeping, not a server feature to build.
- Wii U-specific, platform-built leaderboard classes (source paths resolve to
  `...\Game\ROMWiiU\gaOnline.cpp` / `gaOnlineTester.cpp` / `gaResults.cpp`,
  not dead cross-platform code): `elNetworkLeaderboard::Transaction`,
  `gaOnline::LeaderboardPage`, `RacePointsLeaderboardPage`,
  `TopSpeedLeaderboardPage`, `TimeTrialLeaderboardPage`. `.rodata` has
  hundreds of leaderboard strings: `leaderboardEvent`, `leaderboardDistance`,
  `leaderboardSpeed`, `ViewRankingsButton`, `WorldRankButton`,
  `pfilter_FriendsOnly`, `FriendRivalRankButton`.
- `PRUDPEndPoint`/`PRUDPStream`/`PRUDPMessageV0`/`V1`/`StreamManager`/
  `BackEndServices`, including `s_szSandboxAccessKey`/`SetSandboxAccessKey`/
  `GetSandboxAccessKey` symbols - confirms the access-key plumbing exists,
  but **no literal access-key value, hostname, or game-server-ID string
  exists anywhere in `.rodata`/`.data`/`.strtab`** (no `NASC`, `nasc`,
  `nintendowifi`, `account.nintendo.net` literal strings either) - these are
  supplied at runtime, not compiled in. Same situation as every sibling
  server in this project.

### Matchmaking-adjacent code (stage-5 stretch goal, not day-1 scope)

`wuNetworkOnline::createGame`/`getGameList`/`startMatch` (with
`matchmakingSession` types) and `nn::nex::NintendoPresence::
SetMatchmakeSystemType` are real, Wii U-specific compiled code (source path
`EngineObj\ROMWiiU\wuNetworkOnline.cpp`) - not leftover/dead code. But their
reachability in the shipped retail build is unproven. One secondary web
source (playbite.com, low reliability) claims the game has no online
multiplayer, only local ("Hot Seat Mode") - overstated given the leaderboard
evidence above, but a genuine open question for the matchmaking-specific
code. Investigate as a follow-up once Ranking is console-verified.

### No DataStore evidence

Ghost/replay data (`ghosts/%s/ghost_%03d.bin`, `GhostDownloadIcon`,
`DeleteGhostsButton`) exists in strings, but **no `GhostUpload`/
`GhostDownload`/`GhostTransaction` network symbols were found** - ghosts
are very likely stored locally, not server-synced. Explicit design decision:
**no DataStore/S3 backend in this repo**, unlike Xenoblade Chronicles X.

## Regions & Title IDs

| Region | Product code | Title ID | Source |
|---|---|---|---|
| USA | WUP-P-AHWE | `0005000010143300` | wiiubrew.org Title Database |
| Europe/PAL (covers Australia) | WUP-P-AHWP | `0005000010145100` | wiiubrew.org Title Database |

No Japan release exists. Released September 17, 2013 (USA) and September 20,
2013 (Europe); Australia October 2, 2013 (same PAL build as Europe).
`gametdb.com` blocks direct fetches (HTTP 403) - the Title IDs above came
from wiiubrew.org's Title Database instead, cross-referenced against
GameTDB's game-code listing (`AHWEWR`/`AHWPWR`) via an archived snapshot.

### Game server ID - CONFIRMED `10143300` (EUR console, 2026-09-14)

wsc-account logged a real console request:
`nex_token?game_server_id=10143300 (title=0005000010145100)`. The EUR title
uses the ID derived from the USA title, i.e. one shared ID across regions.
`10145100` is unconfirmed and only kept as a fallback route. Error 102-2482
(INVALID_GAME_SERVER_ID) was caused by the legacy-tls nginx map
(`/opt/wii-sports-club/deploy/legacy-tls/nginx.conf`, `$account_upstream`)
having no entry for this ID, so the request fell through to wsc-account ->
real Pretendo. Fix: `10143300 127.0.0.1:8118` in that map.

### (original hypothesis, superseded above)

`globals/config.go`'s `GameServerIDUSA = "10143300"` and
`GameServerIDEUR = "10145100"` are the low 32 bits of each region's Title ID
- the pattern that held for this project's **Art of Balance** server
(Title ID `0005000010135000` -> game server ID `10135000`, console-confirmed).
**This is a hypothesis, not a confirmed fact for this title.** Two other
multi-region titles in this project (Xenoblade Chronicles X, and Art of
Balance itself, which is nominally multi-region even though its ID turned
out to be single/shared) ended up needing only **one** shared game server ID
across regions rather than one per region. Console-test both region IDs
independently; if the second turns out redundant once the first is
confirmed working, drop it and correct this note rather than leaving two
speculative `PN_GAME_SERVERS` entries in permanently.

## No public NEX database entry

Checked and confirmed empty for this title:

- Kinnay's [`nexwiiu.json`](https://kinnay.github.io/data/nexwiiu.json) - 41
  other Wii U titles listed (Mario Kart 8, Splatoon, Pikmin 3, Smash Wii U,
  etc.), no Hot Wheels / World's Best Driver entry. Schema note: one record
  per game title (not per-region) - fields `access_key`, `game_server_id`
  (decimal), `title_id` (decimal), `branch`, `build`.
  No example in this dataset shows a title split into multiple regional
  records, so the schema itself gives no signal either way on the
  one-ID-vs-two-ID question above.
- PretendoNetwork's `nex-protocols-go`/`nex-protocols-common-go`/`nex-go`
  repos - generic protocol/server libraries, no per-game data at all (that
  lives in deployment configs, not these repos).
- GBATemp / wiiubrew / GitHub search for "TeamHotWheels" or
  "worldsbestdriver" tied to NEX/PRUDP data - nothing.

This is first-party reverse engineering with no public starting point,
unlike every other server in this project.

## Access key - CONFIRMED `783a01d2` (2026-09-18); NEX version still open

No literal access key exists in the binary (see "Linked NEX surface" above)
and no public database documents it. Recovery workflow (same one that
recovered Art of Balance's `96900116`):

1. Deploy this server (with the placeholder `AccessKey = "00000000"`) so
   DNS/routing exists.
2. Register both `PN_GAME_SERVERS` hypotheses on the account server pointing
   at this server's host:port.
3. Get a real console (DNS-redirected to Protarium) to attempt an online
   connection with an actual cartridge/eShop copy of the game.
4. `tcpdump` the PRUDPv1 SYN/CONNECT packet.
5. Run PretendoNetwork's `access-key-extractor -bruteforce -packet=<hex>`
   (ROM-scan mode alone is not reliable - it returned only wrong candidates
   for Art of Balance).
6. Wire the recovered key into `globals/config.go`, redeploy, and iterate on
   `LegacyConnectionSignature`/`UseStructureHeader`/the NEX version pin
   against the real capture until `LoginEx` -> `RequestTicket` -> secure
   `Register` succeeds - the same process every sibling server in this
   project went through.

Step 5 was run against a real captured EUR SYN (signature
`37766fec6357614fb200beab21dc05be`, 24-byte options variant) over the full
8-hex keyspace - exactly one key matched, `783a01d2`, now wired into
`globals/config.go`. The NEX version and wire-format flags below remain
unconfirmed until that same SYN/CONNECT exchange is replayed end to end.

`globals/config.go`'s `NEXMajor/Minor/Patch = 3.4.0` and both
`LegacyConnectionSignature = true` / `UseStructureHeader = false` are
starting guesses (matching this project's other 2013-era titles, Trine 2 and
Sochi 2014), **not confirmed** - expect to revise once a real capture exists,
the same way Xenoblade Chronicles X and Hyrule Warriors both had to.

## Hardware status

**Not yet tested against real hardware.** Nothing below is console-verified.
