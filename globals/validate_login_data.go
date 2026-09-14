package globals

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	account_management_types "github.com/PretendoNetwork/nex-protocols-go/v2/account-management/types"
	ticket_granting_types "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting/types"
)

// ValidateLoginData is the ticket-granting hook that checks the login token
// the client presents. In LocalAuthMode there is no account server issuing
// tokens, so it is accepted unconditionally; otherwise it must decrypt and
// validate against NEXTokenAESKey.
//
// nex-protocols-common-go v2.6.1 replaced its old local AES-key validator
// (ValidatePretendoLoginData, used by every other server in this project)
// with a gRPC call to a separate account service - a heavier piece of
// infrastructure this project's existing token-issuer pattern
// (Protarium-Network/account-server) does not run. So this is the same
// validator logic, ported in directly rather than depending on a removed
// upstream helper: AES-CBC decrypt with the shared secret, CRC32 checksum,
// PID match, expiry and ban checks - identical wire format, same secret as
// every sibling server's account-server deployment.
func ValidateLoginData(pid types.PID, loginData types.DataHolder) *nex.Error {
	if LocalAuthMode {
		return nil
	}

	return validatePretendoLoginData(pid, loginData, NEXTokenAESKey)
}

func validatePretendoLoginData(pid types.PID, loginData types.DataHolder, aesKey []byte) *nex.Error {
	var tokenBase64 string

	loginDataType := loginData.Object.DataObjectID().(types.String)

	switch loginDataType {
	case "NintendoLoginData": // Wii U
		nintendoLoginData := loginData.Object.Copy().(ticket_granting_types.NintendoLoginData)
		tokenBase64 = string(nintendoLoginData.Token)
	case "AccountExtraInfo": // 3DS
		accountExtraInfo := loginData.Object.Copy().(account_management_types.AccountExtraInfo)
		tokenBase64 = string(accountExtraInfo.NEXToken)
		tokenBase64 = strings.ReplaceAll(tokenBase64, ".", "+")
		tokenBase64 = strings.ReplaceAll(tokenBase64, "-", "/")
		tokenBase64 = strings.ReplaceAll(tokenBase64, "*", "=")
	case "AuthenticationInfo": // 3DS / Wii U
		authenticationInfo := loginData.Object.Copy().(ticket_granting_types.AuthenticationInfo)
		tokenBase64 = string(authenticationInfo.Token)
		tokenBase64 = strings.ReplaceAll(tokenBase64, ".", "+")
		tokenBase64 = strings.ReplaceAll(tokenBase64, "-", "/")
		tokenBase64 = strings.ReplaceAll(tokenBase64, "*", "=")
	default:
		Logger.Errorf("Invalid loginData data type %s!", loginDataType)
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, fmt.Sprintf("Invalid loginData data type %s!", loginDataType))
	}

	encryptedToken, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		Logger.Error(err.Error())
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, err.Error())
	}

	decryptedToken, nexError := decryptPretendoToken(encryptedToken, aesKey)
	if nexError != nil {
		return nexError
	}

	if decryptedToken.TokenType != 3 { // NEX token
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Invalid token type")
	}

	expireTime := time.Unix(int64(decryptedToken.ExpireTime/1000), 0) // milliseconds
	if expireTime.Before(time.Now()) {
		return nex.NewError(nex.ResultCodes.Authentication.TokenExpired, "Token expired")
	}

	if types.NewPID(uint64(decryptedToken.UserPID)) != pid {
		return nex.NewError(nex.ResultCodes.Authentication.PrincipalIDUnmatched, fmt.Sprintf("Account %d expected, got %d", pid, decryptedToken.UserPID))
	}

	if decryptedToken.AccessLevel < 0 {
		return nex.NewError(nex.ResultCodes.RendezVous.AccountDisabled, fmt.Sprintf("Account %d is banned", decryptedToken.UserPID))
	}

	return nil
}

// pretendoToken is the Pretendo-specific NEX token format.
type pretendoToken struct {
	SystemType  uint8
	TokenType   uint8
	UserPID     uint32
	ExpireTime  uint64
	TitleID     uint64
	AccessLevel int8
}

func decryptPretendoToken(encryptedToken []byte, aesKey []byte) (*pretendoToken, *nex.Error) {
	if len(encryptedToken) < 4 {
		return nil, nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Token size is too small")
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, err.Error())
	}

	expectedChecksum := binary.BigEndian.Uint32(encryptedToken[0:4])
	encryptedBody := encryptedToken[4:]

	if len(encryptedBody)%aes.BlockSize != 0 {
		return nil, nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, fmt.Sprintf("Encrypted body has invalid size %d", len(encryptedBody)))
	}

	decrypted := make([]byte, len(encryptedBody))
	iv := make([]byte, 16)
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(decrypted, encryptedBody)

	paddingSize := int(decrypted[len(decrypted)-1])
	if paddingSize < 0 || paddingSize >= len(decrypted) {
		return nil, nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, fmt.Sprintf("Invalid padding size %d for token %x", paddingSize, encryptedToken))
	}
	decrypted = decrypted[:len(decrypted)-paddingSize]

	calculatedChecksum := crc32.ChecksumIEEE(decrypted)
	if expectedChecksum != calculatedChecksum {
		return nil, nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Checksum did not match. Failed decrypt. Are you using the right key?")
	}

	token := &pretendoToken{}
	if err := binary.Read(bytes.NewReader(decrypted), binary.LittleEndian, token); err != nil {
		return nil, nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, err.Error())
	}

	return token, nil
}
