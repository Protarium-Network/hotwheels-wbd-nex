# Hot Wheels: World's Best Driver (Wii U) — NEX server

A preservation-oriented NEX server for the Wii U title **Hot Wheels: World's
Best Driver** (Firebrand Games / Mattel / Warner Bros. Interactive, 2013;
internal project name "TeamHotWheels"). It speaks the game's PRUDP
authentication and secure protocols so its **leaderboards** (Race Points, Top
Speed, Time Trial, Career, with friends/rivals filters) work again after the
official servers went away.

Built on the [Pretendo Network](https://github.com/PretendoNetwork) NEX
libraries (`nex-go`, `nex-protocols-go`, `nex-protocols-common-go`) — the
same stack as this org's other servers (Xenoblade Chronicles X, Hyrule
Warriors, Trine 2, Mario & Sonic, Sonic & All-Stars Racing Transformed,
Mario Tennis).

**Unlike every sibling server in this project, no public NEX database
documents this title** — the game server ID and access key below were
recovered from a real console capture; the NEX version is still an open
placeholder. The full evidence trail is in [RECON.md](RECON.md); per-protocol
status is in [PROTOCOL_COVERAGE.md](PROTOCOL_COVERAGE.md).

## Recovered configuration

| Field | Value | Source |
|---|---|---|
| Game server ID | `10143300` — **confirmed** | a real EUR console's `nex_token` request; shared across both regions (see RECON.md) |
| Game server ID (EUR/AUS fallback route) | `10145100` — unverified hypothesis, kept only as a fallback | low 32 bits of Title ID `0005000010145100` |
| Access key | `783a01d2` — **confirmed** | brute-forced against the HMAC-MD5 signature of a real captured PRUDPv1 SYN from an EUR console — see RECON.md |
| NEX version | `3.4.0` — placeholder guess | closest-era match to this project's other 2013 titles (Trine 2, Sochi 2014) |

The retail RPX (`Game.rpx`) statically links NEX (no `nn_nex` RPL import,
same as every sibling title) and carries real, substantial Ranking-protocol
code (`nn::nex::RankingClient`, Wii U-specific leaderboard classes) — see
RECON.md for the full string/symbol evidence.

### Two regions, one confirmed ID

Hot Wheels: World's Best Driver shipped in the USA and Europe (which also
covers the Australian release) — no Japan release exists. A real EUR console
was confirmed to send the USA-derived ID (`10143300`), matching the pattern
every other multi-region title in this project ended up needing — one shared
ID. `10145100` is kept registered only as an unconfirmed fallback route. See
RECON.md's "Game server ID" section.

## Scope

- **Ticket Granting** — login / secure-server handoff.
- **Secure Connection**, **Utility** — baseline secure-endpoint handshake.
- **Ranking** — the confirmed core feature. Global/Own/Friends/Near/
  Friend-Near leaderboard queries, score upload, common data, plus
  `GetApproxOrder`/`GetStats`/`ChangeAttributes`/`DeleteScore` (hand-written —
  no common-go implementation exists for these four).
- **Health + Monitoring** — liveness checks.
- **AccountManagement, RemoteLogDevice, Subscription** — registered with no
  handlers; every method already answers `NotImplemented` on its own.

**Deliberately not registered yet:** MatchMaking/NAT Traversal (real but
unproven-reachable `wuNetworkOnline` code exists — stage-5 follow-up) and
DataStore (no evidence at all; ghosts appear to be stored locally). See
PROTOCOL_COVERAGE.md.

```
        console                       this server
           │
           │  ── LoginEx ─────────────▶  authentication server  :27600
           │  ◀─ Kerberos ticket +
           │     secure server address
           │
           │  ── ticket, RegisterEx ──▶  secure server          :27601
           │  ── GetRanking / UploadScore ▶  leaderboards        → PostgreSQL
```

## Database

One PostgreSQL database, `ranking` schema only (`ranking.scores` +
`ranking.common_data`) — see `database/init_postgres.go`. No DataStore or
matchmaking schema, unlike most sibling servers.

## Running

### Local preservation mode

```bash
cp .env.example .env                     # PN_HWWBD_LOCAL_MODE=1 by default
cp settings.example.json settings.json   # add your console's PID + NEX password
docker compose up --build
```

In local mode there is no account server: player NEX passwords come from
`settings.json` and the login token is accepted unconditionally. Use it only
on an isolated network. Set `PN_HWWBD_SECURE_HOST` to the LAN IP the console
can reach this machine on (not `localhost`, unless the client runs here too).

### Shared mode

Set `PN_HWWBD_LOCAL_MODE` to anything but `1` and provide
`PN_HWWBD_NEX_TOKEN_AES_KEY` (64 hex chars) and `PN_HWWBD_NEX_PASSWORD_SECRET`
(≥32 bytes hex), both matching your account server. Login tokens are then
decrypted and validated, and each player's NEX password is derived as
`HMAC-SHA256(secret, pid)`.

### Without Docker

```bash
export PN_HWWBD_AUTH_PORT=27600 PN_HWWBD_SECURE_PORT=27601
export PN_HWWBD_SECURE_HOST=<LAN-IP-of-this-machine>
export PN_HWWBD_POSTGRES_URI='postgres://hwwbd:hwwbd@localhost:5432/hwwbd?sslmode=disable'
export PN_HWWBD_LOCAL_MODE=1
go build -o hwwbd-nex . && ./hwwbd-nex
```

### Pointing a console at it

On a `Protarium-Network/account-server` deployment, register **both** region
hypotheses pointing at this server, each restricted to its own title so a
console never gets routed to the wrong one:

```
PN_GAME_SERVERS=10143300=<host>:27600:0005000010143300;10145100=<host>:27600:0005000010145100
```

`PN_HWWBD_SECURE_HOST` **must be short** (~15 chars) — the retail binary
truncates it into a fixed-size buffer.

## Hardware status

**Not yet tested against real hardware.** The game server ID and access key
are confirmed from a real console capture, but the NEX version and wire
settings (`LegacyConnectionSignature`, `UseStructureHeader`) are still
unverified placeholders — see RECON.md for the exact recon workflow this
project used to recover the same values for Art of Balance and every other
sibling server.

## License

AGPL-3.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE). No proprietary
Nintendo or Firebrand Games code or assets are included.
