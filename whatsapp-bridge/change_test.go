package main

import (
	"testing"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// The secret-encrypted edit path needs a live client to decrypt, so only the
// protocol-message paths and the "not a change" cases are covered here.
func TestMessageChange(t *testing.T) {
	key := &waProto.MessageKey{ID: ptr("ORIGINAL")}
	cases := []struct {
		name                      string
		msg                       *waProto.Message
		kind, originalID, newText string
		ok                        bool
	}{
		{
			"revoke",
			&waProto.Message{ProtocolMessage: &waProto.ProtocolMessage{
				Type: waProto.ProtocolMessage_REVOKE.Enum(), Key: key,
			}},
			"deleted", "ORIGINAL", "", true,
		},
		{
			"edit",
			&waProto.Message{ProtocolMessage: &waProto.ProtocolMessage{
				Type: waProto.ProtocolMessage_MESSAGE_EDIT.Enum(), Key: key,
				EditedMessage: &waProto.Message{Conversation: ptr("fixed")},
			}},
			"edited", "ORIGINAL", "fixed", true,
		},
		{
			"other protocol message is not stored",
			&waProto.Message{ProtocolMessage: &waProto.ProtocolMessage{
				Type: waProto.ProtocolMessage_HISTORY_SYNC_NOTIFICATION.Enum(),
			}},
			"", "", "", true,
		},
		{
			// GetType() on a missing protocol message is REVOKE (0), so this guards the nil check
			"plain text is not a change",
			&waProto.Message{Conversation: ptr("hello")},
			"", "", "", false,
		},
		{
			"secret-encrypted poll edit is not a message change",
			&waProto.Message{SecretEncryptedMessage: &waProto.SecretEncryptedMessage{
				SecretEncType: waE2E.SecretEncryptedMessage_POLL_EDIT.Enum(), TargetMessageKey: key,
			}},
			"", "", "", false,
		},
	}
	logger := waLog.Noop
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			kind, originalID, newText, ok := messageChange(nil, &events.Message{Message: c.msg}, logger)
			if kind != c.kind || originalID != c.originalID || newText != c.newText || ok != c.ok {
				t.Errorf("got (%q, %q, %q, %v), want (%q, %q, %q, %v)",
					kind, originalID, newText, ok, c.kind, c.originalID, c.newText, c.ok)
			}
		})
	}
}

func TestMessageFields(t *testing.T) {
	msg := &waProto.Message{
		ReactionMessage:    &waProto.ReactionMessage{Text: ptr("👍")},
		MessageContextInfo: &waProto.MessageContextInfo{},
	}
	got := messageFields(msg)
	if got != "messageContextInfo, reactionMessage" && got != "reactionMessage, messageContextInfo" {
		t.Errorf("got %q", got)
	}
}
