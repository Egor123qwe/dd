package response

import (
	"encoding/json"

	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/message"
)

const (
	err = "error"
)

func BaseError(mt, status, msg string) []byte {
	resp := &ErrorRespnse{
		Type: mt,
		Meta: message.Meta{
			Status: status,
			Err: &message.Err{
				Code:    err,
				Message: msg,
			},
		},
	}

	respBytes, _ := json.Marshal(resp)
	return respBytes
}

func ErrorWS(mt, connID, messageID, status, msg string) []byte {
	resp := &ErrorRespnse{
		Type: mt,
		Meta: message.Meta{
			Status:    status,
			MessageID: messageID,
			Err: &message.Err{
				Code:    err,
				Message: msg,
			},
			Conn: message.Connection{
				ConnectionID: connID,
				Type:         "conn_id",
			},
		},
	}

	respBytes, _ := json.Marshal(resp)
	return respBytes
}

func ErrorSession(mt, sessionID, connID, messageID, status, msg string) []byte {
	resp := &ErrorRespnse{
		Type: mt,
		Meta: message.Meta{
			Status:    status,
			MessageID: messageID,
			Err: &message.Err{
				Code:    err,
				Message: msg,
			},
			Conn: message.Connection{
				ConnectionID: connID,
				Type:         "conn_id",
			},
		},
		Content: ContentErrorResponse{
			SessionID: sessionID,
		},
	}

	respBytes, _ := json.Marshal(resp)
	return respBytes
}

type ErrorRespnse struct {
	Type    string               `json:"type"`
	Meta    message.Meta         `json:"meta"`
	Content ContentErrorResponse `json:"content"`
}

type ContentErrorResponse struct {
	SessionID string `json:"session_id"`
}
