package nex

import (
	"os"
	"strconv"

	nex "github.com/PretendoNetwork/nex-go/v2"

	"github.com/Protarium-Network/hotwheels-wbd-nex/globals"
)

var secureTracer = newTracer("secure")

func StartSecureServer() {
	globals.SecureServer = nex.NewPRUDPServer()

	// See authentication.go - both UNVERIFIED, no packet capture of this
	// title exists yet.
	globals.SecureServer.PRUDPV1Settings.LegacyConnectionSignature = true
	globals.SecureServer.ByteStreamSettings.UseStructureHeader = false

	globals.SecureEndpoint = nex.NewPRUDPEndPoint(1)
	globals.SecureEndpoint.IsSecureEndPoint = true
	globals.SecureEndpoint.ServerAccount = globals.SecureServerAccount
	globals.SecureEndpoint.AccountDetailsByPID = globals.AccountDetailsByPID
	globals.SecureEndpoint.AccountDetailsByUsername = globals.AccountDetailsByUsername
	globals.SecureServer.BindPRUDPEndPoint(globals.SecureEndpoint)

	globals.SecureServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(globals.NEXMajor, globals.NEXMinor, globals.NEXPatch))
	globals.SecureServer.AccessKey = globals.AccessKey

	secureTracer.attachLogging(globals.SecureEndpoint)

	registerSecureProtocols()

	secureTracer.attachFallback(globals.SecureEndpoint)

	port, _ := strconv.Atoi(os.Getenv("PN_HWWBD_SECURE_PORT"))
	claimPort("Secure server", port)
	globals.Logger.Successf("[HWWBD] Secure server listening on UDP %d", port)
	globals.SecureServer.Listen(port)
}
