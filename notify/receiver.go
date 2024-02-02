package notify

import (
	"github.com/go-lark/lark"
)

/*
	for msg receiver
*/

// ReceiverType  ...
type ReceiverType string

var (
	ReceiverEmail  ReceiverType = "email"
	ReceiverChatID ReceiverType = "chatID"
	ReceiverUserID ReceiverType = "userID"
	ReceiverOpenID ReceiverType = "openID"
)

// EmailReceiver create a receiver from email.
func EmailReceiver(email string) *Receiver {
	return &Receiver{ID: email, Type: ReceiverEmail}
}

// ChatReceiver create a receiver from a chat id.
func ChatReceiver(chatID string) *Receiver {
	return &Receiver{ID: chatID, Type: ReceiverChatID}
}

// Receiver represent a lark message receiver
type Receiver struct {
	ID   string // receiver ident
	Type ReceiverType
}

func wrapReceiver(msg *lark.MsgBuffer, receiver *Receiver) *lark.MsgBuffer {
	switch receiver.Type {
	case ReceiverChatID:
		msg = msg.BindChatID(receiver.ID)
	case ReceiverUserID:
		msg = msg.BindUserID(receiver.ID)
	case ReceiverOpenID:
		msg = msg.BindOpenID(receiver.ID)
	default:
		msg = msg.BindEmail(receiver.ID)
	}
	return msg
}
