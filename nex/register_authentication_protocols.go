package nex

import (
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_ticket_granting "github.com/PretendoNetwork/nex-protocols-common-go/v2/ticket-granting"
	ticket_granting "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting"

	"github.com/Protarium-Network/hotwheels-wbd-nex/globals"
)

func registerAuthenticationProtocols() {
	ticketGrantingProtocol := ticket_granting.NewProtocol()
	globals.AuthenticationEndpoint.RegisterServiceProtocol(ticketGrantingProtocol)
	authTracer.register(ticket_granting.ProtocolID)

	commonTicketGranting := common_ticket_granting.NewCommonProtocol(ticketGrantingProtocol)
	commonTicketGranting.ValidateLoginData = globals.ValidateLoginData
	commonTicketGranting.SecureStationURL = secureStationURLFromEnv()
	commonTicketGranting.BuildName = types.NewString("")
	commonTicketGranting.SecureServerAccount = globals.SecureServerAccount
}
