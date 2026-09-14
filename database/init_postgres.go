package database

import (
	"os"

	"github.com/Protarium-Network/hotwheels-wbd-nex/globals"
)

// initPostgres creates the schema this server needs on first run. Every
// statement is idempotent, so it is safe to run on every start.
//
// Only Ranking (leaderboards) is created - Game.rpx shows no evidence of
// DataStore usage (ghosts are stored locally, no GhostUpload/GhostDownload
// symbols) or a working Matchmaking session layer, so unlike most sibling
// servers there is no datastore/matchmaking/tracking schema here. See
// RECON.md.
func initPostgres() {
	mustExec := func(label, query string) {
		if _, err := Postgres.Exec(query); err != nil {
			globals.Logger.Criticalf("%s: %s", label, err.Error())
			os.Exit(1)
		}
	}

	mustExec("ranking schema", `CREATE SCHEMA IF NOT EXISTS ranking`)
	mustExec("ranking.scores", `CREATE TABLE IF NOT EXISTS ranking.scores (
		pid         numeric(20) NOT NULL,
		unique_id   numeric(20) NOT NULL DEFAULT 0,
		category    bigint      NOT NULL,
		score       bigint      NOT NULL,
		order_by    smallint    NOT NULL DEFAULT 0,
		groups      bytea       NOT NULL DEFAULT ''::bytea,
		param       numeric(20) NOT NULL DEFAULT 0,
		update_time timestamp   NOT NULL DEFAULT (now() at time zone 'utc'),
		PRIMARY KEY (pid, unique_id, category)
	)`)
	mustExec("ranking.scores index", `CREATE INDEX IF NOT EXISTS scores_category_score_idx ON ranking.scores (category, score)`)

	mustExec("ranking.common_data", `CREATE TABLE IF NOT EXISTS ranking.common_data (
		unique_id numeric(20) NOT NULL,
		pid       numeric(20) NOT NULL,
		data      bytea       NOT NULL,
		PRIMARY KEY (unique_id, pid)
	)`)

	globals.Logger.Success("Postgres schema ready (ranking only)")
}
