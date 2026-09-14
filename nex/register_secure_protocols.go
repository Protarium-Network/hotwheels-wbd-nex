package nex

import (
	common_secure "github.com/PretendoNetwork/nex-protocols-common-go/v2/secure-connection"
	common_utility "github.com/PretendoNetwork/nex-protocols-common-go/v2/utility"

	account_management "github.com/PretendoNetwork/nex-protocols-go/v2/account-management"
	health "github.com/PretendoNetwork/nex-protocols-go/v2/health"
	monitoring "github.com/PretendoNetwork/nex-protocols-go/v2/monitoring"
	remote_log_device "github.com/PretendoNetwork/nex-protocols-go/v2/remote-log-device"
	secure_connection "github.com/PretendoNetwork/nex-protocols-go/v2/secure-connection"
	subscription "github.com/PretendoNetwork/nex-protocols-go/v2/subscription"
	utility "github.com/PretendoNetwork/nex-protocols-go/v2/utility"

	"github.com/Protarium-Network/hotwheels-wbd-nex/globals"
)

// registerSecureProtocols wires up what Game.rpx's RankingClient/leaderboard
// evidence supports (see RECON.md): the baseline secure handshake and
// Ranking. wuNetworkOnline's matchmaking-adjacent code
// (createGame/getGameList/startMatch) is real but its reachability in the
// shipped retail build is unproven - MatchMaking/MatchMakingExt/NAT Traversal
// are deliberately NOT registered yet; investigate and wire them in as a
// stage-5 follow-up once Ranking is console-verified (see RECON.md).
// DataStore/Messaging have no evidence at all (ghosts are local-only, no
// GhostUpload/GhostDownload symbols) and are skipped entirely, unlike
// Xenoblade.
func registerSecureProtocols() {
	registerSecureConnection()
	registerUtility()
	registerRanking()
	registerLiveness()
	registerStubs()
}

func registerSecureConnection() {
	secureProtocol := secure_connection.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(secureProtocol)
	secureTracer.register(secure_connection.ProtocolID)

	secureCommon := common_secure.NewCommonProtocol(secureProtocol)
	secureCommon.EnableInsecureRegister()
}

func registerUtility() {
	utilityProtocol := utility.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(utilityProtocol)
	secureTracer.register(utility.ProtocolID)

	common_utility.NewCommonProtocol(utilityProtocol)
}

func registerLiveness() {
	monitoringProtocol := monitoring.NewProtocol()
	monitoringProtocol.SetHandlerPingDaemon(pong(monitoring.ProtocolID, monitoring.MethodPingDaemon))
	globals.SecureEndpoint.RegisterServiceProtocol(monitoringProtocol)
	secureTracer.register(monitoring.ProtocolID)

	healthProtocol := health.NewProtocol()
	healthProtocol.SetHandlerPingDaemon(pong(health.ProtocolID, health.MethodPingDaemon))
	healthProtocol.SetHandlerPingDatabase(pong(health.ProtocolID, health.MethodPingDatabase))
	globals.SecureEndpoint.RegisterServiceProtocol(healthProtocol)
	secureTracer.register(health.ProtocolID)
}

// registerStubs registers the three linked-but-unimplemented protocols this
// project's servers consistently register defensively - every method here
// already answers NotImplemented on its own when the handler is nil.
func registerStubs() {
	accountManagementProtocol := account_management.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(accountManagementProtocol)
	secureTracer.register(account_management.ProtocolID)

	remoteLogDeviceProtocol := remote_log_device.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(remoteLogDeviceProtocol)
	secureTracer.register(remote_log_device.ProtocolID)

	subscriptionProtocol := subscription.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(subscriptionProtocol)
	secureTracer.register(subscription.ProtocolID)
}
