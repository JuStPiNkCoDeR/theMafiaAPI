package main

import "mafia/sockets/lib"

/*
=================
| Secure Socket |
=================
*/
// Send servers' public keys to client
func getServerKeys(eventName string, ss *SecureSocket) {
	triggerSocketEvent(fmtLogger, eventName, ss.SimpleSocket)

	if err := ss.SendPublicKeys(); err != nil {
		errorSocketEvent(fmtLogger, eventName, err, ss.SimpleSocket)
		return
	}

	successfulSocketEvent(fmtLogger, eventName, ss.SimpleSocket)
}

func setClientsKeys(incomingMessageMap map[string]interface{}, eventName string, ss *SecureSocket) bool {
	triggerSocketEvent(fmtLogger, eventName, ss.SimpleSocket)

	incomingData := incomingMessageMap["data"].(map[string]interface{})

	// Parse OAEP PEM data
	if pemOAEP, ok := incomingData["oaep"]; ok {
		switch typedPEM := pemOAEP.(type) {
		case string:
			if err := ss.rsa.ImportKey(typedPEM, true); err != nil {
				errorSocketEvent(
					fmtLogger, eventName,
					&lib.StackError{
						ParentError: err,
						Message:     "Error on import OAEP key",
					}, ss.SimpleSocket)

				return false
			}

			if ss.rsa.ForeignPublicKeyOAEP == nil {
				fmtLogger.Log(lib.Warn, "Foreign OAEP rsa key is nil after import")
			}
		default:
			errorSocketEvent(fmtLogger, eventName, &lib.StackError{
				Message: "PEM string is not string!",
			}, ss.SimpleSocket)

			return false
		}
	}
	// Parse PSS PEM data
	if pemPSS, ok := incomingData["pss"]; ok {
		switch typedPEM := pemPSS.(type) {
		case string:
			if err := ss.rsa.ImportKey(typedPEM, false); err != nil {
				errorSocketEvent(
					fmtLogger, eventName,
					&lib.StackError{
						ParentError: err,
						Message:     "Error on import PSS key",
					}, ss.SimpleSocket)

				return false
			}

			if ss.rsa.ForeignPublicKeyOAEP == nil {
				fmtLogger.Log(lib.Warn, "Foreign PSS rsa key is nil after import")
			}
		default:
			errorSocketEvent(fmtLogger, eventName, &lib.StackError{
				Message: "PEM string is not string!",
			}, ss.SimpleSocket)

			return false
		}
	}

	successfulSocketEvent(fmtLogger, eventName, ss.SimpleSocket)

	return true
}
