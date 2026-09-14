package nex

import (
	nex "github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_ranking "github.com/PretendoNetwork/nex-protocols-common-go/v2/ranking"
	ranking "github.com/PretendoNetwork/nex-protocols-go/v2/ranking"
	ranking_constants "github.com/PretendoNetwork/nex-protocols-go/v2/ranking/constants"
	ranking_types "github.com/PretendoNetwork/nex-protocols-go/v2/ranking/types"

	"github.com/Protarium-Network/hotwheels-wbd-nex/database"
	"github.com/Protarium-Network/hotwheels-wbd-nex/globals"
)

// registerRanking backs the leaderboard pages confirmed in Game.rpx (Race
// Points / Top Speed / Time Trial / Career, with friends/rivals filters) -
// see RECON.md. GetApproxOrder, GetStats, ChangeAttributes and DeleteScore
// have no common-go implementation (confirmed by reading
// nex-protocols-common-go v2.6.1's ranking/protocol.go), so they're wired
// directly onto the raw protocol struct instead - see
// database/ranking_gaps.go.
func registerRanking() {
	rankingProtocol := ranking.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(rankingProtocol)
	secureTracer.register(ranking.ProtocolID)

	rankingCommon := common_ranking.NewCommonProtocol(rankingProtocol)
	rankingCommon.GetRankingsAndCountByCategoryAndRankingOrderParam = database.GetRankingsAndCountByCategoryAndRankingOrderParam
	rankingCommon.GetOwnRankingByCategoryAndRankingOrderParam = database.GetOwnRankingByCategoryAndRankingOrderParam
	rankingCommon.GetFriendsRankingsAndCountByCategoryAndRankingOrderParam = database.GetFriendsRankingsAndCountByCategoryAndRankingOrderParam
	rankingCommon.GetNearbyRankingsAndCountByCategoryAndRankingOrderParam = database.GetNearbyRankingsAndCountByCategoryAndRankingOrderParam
	rankingCommon.GetNearbyFriendsRankingsAndCountByCategoryAndRankingOrderParam = database.GetNearbyFriendsRankingsAndCountByCategoryAndRankingOrderParam
	rankingCommon.GetCommonData = database.GetCommonData
	rankingCommon.UploadCommonData = database.UploadCommonData
	rankingCommon.InsertRankingByPIDAndRankingScoreData = database.InsertRankingByPIDAndRankingScoreData

	rankingProtocol.GetApproxOrder = handleGetApproxOrder
	rankingProtocol.GetStats = handleGetStats
	rankingProtocol.ChangeAttributes = handleChangeAttributes
	rankingProtocol.DeleteScore = handleDeleteScore
}

func handleGetApproxOrder(err error, packet nex.PacketInterface, callID uint32, category types.UInt32, orderParam ranking_types.RankingOrderParam, score types.UInt32, uniqueID types.UInt64, principalID types.PID) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Ranking.InvalidArgument, err.Error())
	}

	endpoint := packet.Sender().Endpoint()

	order, dbErr := database.GetApproxOrder(category, orderParam, score, uniqueID, principalID)
	if dbErr != nil {
		globals.Logger.Errorf("[HWWBD Ranking] GetApproxOrder: %s", dbErr.Error())
		return nil, nex.NewError(nex.ResultCodes.Ranking.Unknown, "get_approx_order_failed")
	}

	stream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
	types.NewUInt32(order).WriteTo(stream)

	response := nex.NewRMCSuccess(endpoint, stream.Bytes())
	response.ProtocolID = ranking.ProtocolID
	response.MethodID = ranking.MethodGetApproxOrder
	response.CallID = callID

	return response, nil
}

func handleGetStats(err error, packet nex.PacketInterface, callID uint32, category types.UInt32, orderParam ranking_types.RankingOrderParam, flags ranking_constants.StatsFlag) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Ranking.InvalidArgument, err.Error())
	}

	endpoint := packet.Sender().Endpoint()

	stats, dbErr := database.GetStats(category, orderParam, flags)
	if dbErr != nil {
		globals.Logger.Errorf("[HWWBD Ranking] GetStats: %s", dbErr.Error())
		return nil, nex.NewError(nex.ResultCodes.Ranking.Unknown, "get_stats_failed")
	}

	stream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
	stats.WriteTo(stream)

	response := nex.NewRMCSuccess(endpoint, stream.Bytes())
	response.ProtocolID = ranking.ProtocolID
	response.MethodID = ranking.MethodGetStats
	response.CallID = callID

	return response, nil
}

func handleChangeAttributes(err error, packet nex.PacketInterface, callID uint32, category types.UInt32, changeParam ranking_types.RankingChangeAttributesParam, uniqueID types.UInt64) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Ranking.InvalidArgument, err.Error())
	}

	endpoint := packet.Sender().Endpoint()
	pid := packet.Sender().PID()

	if dbErr := database.ChangeAttributes(pid, category, changeParam, uniqueID); dbErr != nil {
		globals.Logger.Errorf("[HWWBD Ranking] ChangeAttributes: %s", dbErr.Error())
		return nil, nex.NewError(nex.ResultCodes.Ranking.Unknown, "change_attributes_failed")
	}

	response := nex.NewRMCSuccess(endpoint, nil)
	response.ProtocolID = ranking.ProtocolID
	response.MethodID = ranking.MethodChangeAttributes
	response.CallID = callID

	return response, nil
}

func handleDeleteScore(err error, packet nex.PacketInterface, callID uint32, category types.UInt32, uniqueID types.UInt64) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Ranking.InvalidArgument, err.Error())
	}

	endpoint := packet.Sender().Endpoint()
	pid := packet.Sender().PID()

	if dbErr := database.DeleteScore(pid, category, uniqueID); dbErr != nil {
		globals.Logger.Errorf("[HWWBD Ranking] DeleteScore: %s", dbErr.Error())
		return nil, nex.NewError(nex.ResultCodes.Ranking.Unknown, "delete_score_failed")
	}

	response := nex.NewRMCSuccess(endpoint, nil)
	response.ProtocolID = ranking.ProtocolID
	response.MethodID = ranking.MethodDeleteScore
	response.CallID = callID

	return response, nil
}
