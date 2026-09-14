// Package nex is the Hot Wheels: World's Best Driver (Wii U) NEX server. See
// RECON.md for the full reverse-engineering evidence trail behind every
// constant and wire setting below - unlike most sibling servers in this
// project, none of it is confirmed against a real packet capture yet.
package nex

import (
	"fmt"
	"os"
	"strconv"

	nex "github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"

	"github.com/Protarium-Network/hotwheels-wbd-nex/globals"
)

var authTracer = newTracer("auth")

func StartAuthenticationServer() {
	globals.AuthenticationServer = nex.NewPRUDPServer()

	// UNVERIFIED placeholder - see globals/config.go and RECON.md. No packet
	// capture of this title exists yet; started at the setting this
	// project's other 2013-era titles (Trine 2, Sochi 2014) confirmed on
	// hardware, as the closest-era guess.
	globals.AuthenticationServer.PRUDPV1Settings.LegacyConnectionSignature = true

	globals.AuthenticationEndpoint = nex.NewPRUDPEndPoint(1)
	globals.AuthenticationEndpoint.ServerAccount = globals.AuthenticationServerAccount
	globals.AuthenticationEndpoint.AccountDetailsByPID = globals.AccountDetailsByPID
	globals.AuthenticationEndpoint.AccountDetailsByUsername = globals.AccountDetailsByUsername
	globals.AuthenticationServer.BindPRUDPEndPoint(globals.AuthenticationEndpoint)

	globals.AuthenticationServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(globals.NEXMajor, globals.NEXMinor, globals.NEXPatch))
	globals.AuthenticationServer.AccessKey = globals.AccessKey
	// UNVERIFIED placeholder, must stay the same value on both endpoints -
	// see nex/secure.go and RECON.md.
	globals.AuthenticationServer.ByteStreamSettings.UseStructureHeader = false

	authTracer.attachLogging(globals.AuthenticationEndpoint)
	globals.AuthenticationEndpoint.OnData(func(packet nex.PacketInterface) {
		request := packet.RMCMessage()
		if request == nil {
			return
		}
		// LoginEx - dump the raw RMC parameter bytes so AuthenticationInfo's
		// layout can be confirmed by hand against a real capture.
		if request.ProtocolID == 0x0A && request.MethodID == 0x02 {
			fmt.Printf("[HWWBD Auth] RAW LoginEx params (%d bytes): %x\n", len(request.Parameters), request.Parameters)
		}
	})

	registerAuthenticationProtocols()

	authTracer.attachFallback(globals.AuthenticationEndpoint)

	port, _ := strconv.Atoi(os.Getenv("PN_HWWBD_AUTH_PORT"))
	claimPort("Authentication server", port)
	globals.Logger.Successf("[HWWBD] Authentication server listening on UDP %d", port)
	globals.AuthenticationServer.Listen(port)
}

// secureStationURLFromEnv builds the StationURL the Ticket Granting response
// tells the console to reconnect to for the secure server. Shared with
// register_authentication_protocols.go.
func secureStationURLFromEnv() types.StationURL {
	securePort, _ := strconv.Atoi(os.Getenv("PN_HWWBD_SECURE_PORT"))

	// Must stay short (~15 chars): the retail binary truncates this field
	// into a small fixed-size buffer, so use a bare host, not a subdomain.
	secureHost := os.Getenv("PN_HWWBD_SECURE_HOST")
	if secureHost == "" {
		secureHost = "localhost"
	}

	url := types.NewStationURL("")
	url.SetURLType(constants.StationURLPRUDPS)
	url.SetAddress(secureHost)
	url.SetPortNumber(uint16(securePort))
	url.SetConnectionID(1)
	url.SetPrincipalID(types.NewPID(2))
	url.SetStreamID(1)
	url.SetStreamType(constants.StreamTypeRVSecure)
	url.SetType(uint8(constants.StationURLFlagPublic))

	return url
}
