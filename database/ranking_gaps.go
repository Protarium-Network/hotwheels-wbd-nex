package database

import (
	"fmt"

	"github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/PretendoNetwork/nex-protocols-go/v2/ranking/constants"
	ranking_types "github.com/PretendoNetwork/nex-protocols-go/v2/ranking/types"
)

// GetApproxOrder, GetStats, ChangeAttributes and DeleteScore have no
// implementation in nex-protocols-common-go v2.6.1's ranking.CommonProtocol
// (confirmed by reading its source: only UploadScore, GetRanking,
// GetCommonData, UploadCommonData and the two GetCachedTopXRanking(s)
// variants are wired) - genuinely new work for this project. Response
// shapes below follow the documented NEX Ranking protocol; none of this is
// tuned to a real capture of this title yet - see RECON.md.

// GetApproxOrder returns the rank a hypothetical score would land at,
// without storing it - "approx" because ties are resolved by count rather
// than the tie-break ordering GetRanking itself uses.
func GetApproxOrder(category types.UInt32, orderParam ranking_types.RankingOrderParam, score types.UInt32, uniqueID types.UInt64, principalID types.PID) (uint32, error) {
	orderBy := categoryOrderBy(category)

	comparison := "<"
	if orderBy == constants.OrderByDescending {
		comparison = ">"
	}

	args := []any{uint32(category)}
	where := "category = $1" + groupFilter(orderParam, &args)
	args = append(args, uint32(score))

	query := fmt.Sprintf(`SELECT COUNT(*) + 1 FROM ranking.scores WHERE %s AND score %s $%d`, where, comparison, len(args))

	var order uint32
	if err := Postgres.QueryRow(query, args...).Scan(&order); err != nil {
		return 0, err
	}

	return order, nil
}

// GetStats returns the aggregate stats a client asks for via StatsFlag, in
// the flag's own bit order (Total, Sum, Min, Max, Average).
func GetStats(category types.UInt32, orderParam ranking_types.RankingOrderParam, flags constants.StatsFlag) (ranking_types.RankingStats, error) {
	stats := ranking_types.NewRankingStats()

	args := []any{uint32(category)}
	where := "category = $1" + groupFilter(orderParam, &args)

	query := fmt.Sprintf(`
		SELECT COUNT(*), COALESCE(SUM(score),0), COALESCE(MIN(score),0), COALESCE(MAX(score),0), COALESCE(AVG(score),0)
		FROM ranking.scores WHERE %s`, where)

	var count int64
	var sum, min, max, avg float64
	if err := Postgres.QueryRow(query, args...).Scan(&count, &sum, &min, &max, &avg); err != nil {
		return stats, err
	}

	add := func(flag constants.StatsFlag, value float64) {
		if flags.HasFlag(flag) {
			stats.StatsList = append(stats.StatsList, types.NewDouble(value))
		}
	}

	add(constants.StatsFlagTotal, float64(count))
	add(constants.StatsFlagSum, sum)
	add(constants.StatsFlagMin, min)
	add(constants.StatsFlagMax, max)
	add(constants.StatsFlagAverage, avg)

	return stats, nil
}

// ChangeAttributes updates the group bytes and/or param of an existing
// score row without touching its score, gated by ModificationFlag.
// RankingChangeAttributesParam.Groups carries one new byte value per group
// slot flagged for update (Group0..Group3, in that order) - inferred from
// modification_flag.go's doc comments, not yet confirmed against a real
// capture.
func ChangeAttributes(pid types.PID, category types.UInt32, changeParam ranking_types.RankingChangeAttributesParam, uniqueID types.UInt64) error {
	groupFlags := []constants.ModificationFlag{
		constants.ModificationFlagGroup0,
		constants.ModificationFlagGroup1,
		constants.ModificationFlagGroup2,
		constants.ModificationFlagGroup3,
	}

	var existing []byte
	err := Postgres.QueryRow(`SELECT groups FROM ranking.scores WHERE pid = $1 AND unique_id = $2 AND category = $3`,
		uint64(pid), uint64(uniqueID), uint32(category)).Scan(&existing)
	if err != nil {
		return err
	}

	values := changeParam.Groups
	valueIndex := 0
	for groupIndex, flag := range groupFlags {
		if !changeParam.ModificationFlag.HasFlag(flag) {
			continue
		}
		if valueIndex >= len(values) {
			return fmt.Errorf("ChangeAttributes: ModificationFlag requests group %d but Groups has only %d entries", groupIndex, len(values))
		}
		for len(existing) <= groupIndex {
			existing = append(existing, 0)
		}
		existing[groupIndex] = byte(values[valueIndex])
		valueIndex++
	}

	args := []any{existing}
	setClauses := "groups = $1"
	if changeParam.ModificationFlag.HasFlag(constants.ModificationFlagParam) {
		args = append(args, uint64(changeParam.Param))
		setClauses += fmt.Sprintf(", param = $%d", len(args))
	}

	args = append(args, uint64(pid), uint64(uniqueID), uint32(category))
	query := fmt.Sprintf(`UPDATE ranking.scores SET %s WHERE pid = $%d AND unique_id = $%d AND category = $%d`,
		setClauses, len(args)-2, len(args)-1, len(args))

	_, err = Postgres.Exec(query, args...)
	return err
}

// DeleteScore removes one score row.
func DeleteScore(pid types.PID, category types.UInt32, uniqueID types.UInt64) error {
	_, err := Postgres.Exec(`DELETE FROM ranking.scores WHERE pid = $1 AND unique_id = $2 AND category = $3`,
		uint64(pid), uint64(uniqueID), uint32(category))
	return err
}
