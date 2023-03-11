package main

import (
	"fmt"
	"time"

	log "maunium.net/go/maulogger/v2"

	"go.mau.fi/mautrix-imessage/imessage"
)

type iMessageHandler struct {
	bridge *customBridge
	log    log.Logger
	stop   chan struct{}
}

func NewiMessageHandler(bridge *customBridge) *iMessageHandler {
	return &iMessageHandler{
		bridge: bridge,
		log:    bridge.Log.Sub("iMessage"),
		stop:   make(chan struct{}),
	}
}

func (imh *iMessageHandler) Start() {

	messages := imh.bridge.IM.MessageChan()
	readReceipts := imh.bridge.IM.ReadReceiptChan()
	typingNotifications := imh.bridge.IM.TypingNotificationChan()
	chats := imh.bridge.IM.ChatChan()
	contacts := imh.bridge.IM.ContactChan()
	messageStatuses := imh.bridge.IM.MessageStatusChan()
	backfillTasks := imh.bridge.IM.BackfillTaskChan()
	for {
		start := time.Now()
		var thing string
		select {
		case msg := <-messages:
			imh.HandleMessage(msg)
			thing = "message"
		case rr := <-readReceipts:
			imh.HandleReadReceipt(rr)
			thing = "read receipt"
		case notif := <-typingNotifications:
			imh.HandleTypingNotification(notif)
			thing = "typing notification"
		case chat := <-chats:
			imh.HandleChat(chat)
			thing = "chat"
		case contact := <-contacts:
			imh.HandleContact(contact)
			thing = "contact"
		case status := <-messageStatuses:
			imh.HandleMessageStatus(status)
			thing = "message status"
		case backfillTask := <-backfillTasks:
			imh.HandleBackfillTask(backfillTask)
			thing = "backfill task"
		case <-imh.stop:
			return
		}
		imh.log.Debugfln(
			"Handled %s in %s (queued: %dm/%dr/%dt/%dch/%dct/%ds/%db)",
			thing, time.Since(start),
			len(messages), len(readReceipts), len(typingNotifications), len(chats), len(contacts), len(messageStatuses), len(backfillTasks),
		)
	}
}

const PortalBufferTimeout = 10 * time.Second

func (imh *iMessageHandler) HandleMessage(msg *imessage.Message) {
	chatID := msg.ChatGUID
	threadID := msg.ThreadID
	chat, err := imh.bridge.IM.GetChatInfo(chatID, threadID)
	if err != nil {
		imh.log.Warnln("Failed to get chat info for incoming message:", err)
		return
	}
	channelID := chatID
	channelName := chat.DisplayName
	if channelName == "" {
		channelName = chat.LocalID
	}
	userID := msg.Sender.String()
	if userID == "" {
		return
	}
	username := msg.Sender.LocalID
	contact, _ := imh.bridge.IM.GetContactInfo(msg.Sender.String())
	if contact != nil {
		username = contact.FirstName + " " + contact.LastName
	}
	userMembersID := map[string]string{}
	event := ""
	if chat.IsGroup {
		for _, member := range chat.Members {
			contact, _ := imh.bridge.IM.GetContactInfo(member)
			if contact != nil {
				username := contact.FirstName + " " + contact.LastName
				userMembersID[member] = username
			}
		}

		event = "new_users"
	} else {
		event = "direct_msg"
	}
	// in case contact info can't be fetched , it will populate the map with the sender info
	if len(userMembersID) == 0 {
		userMembersID[userID] = username
	}
	mtbridgeMsg := Message{
		Text:      msg.Text,
		Channel:   channelID,
		Username:  username,
		UserID:    userID,
		Avatar:    "",
		Account:   account,
		Event:     event,
		Protocol:  "api",
		Gateway:   "",
		ParentID:  "",
		Timestamp: time.Time{},
		ID:        chatID,
		Extra:     map[string][]interface{}{},
		ExtraNetworkInfo: ExtraNetworkInfo{
			ActionCommand:  "imessage",
			ChannelId:      channelID,
			ChannelName:    channelName,
			TargetPlatform: "appservice",

			UsersMemberId: userMembersID,
		},
	}
	imh.HandleAttachments(&mtbridgeMsg, msg)
	imh.bridge.HandleRemoteApiSend(mtbridgeMsg)
	fmt.Printf("Received  message: %s , from user %s\n", msg.Text, msg.Sender.LocalID)
}

func (imh *iMessageHandler) HandleMessageStatus(status *imessage.SendMessageStatus) {
	fmt.Printf("Received message status: %+v", status)
}

func (imh *iMessageHandler) HandleReadReceipt(rr *imessage.ReadReceipt) {
	fmt.Printf("Received read receipt: %+v", rr)
}

func (imh *iMessageHandler) HandleTypingNotification(notif *imessage.TypingNotification) {
	fmt.Printf("Received typing notification: %+v", notif)
}

func (imh *iMessageHandler) HandleChat(chat *imessage.ChatInfo) {
	fmt.Printf("Received chat info: %+v", chat)
}

func (imh *iMessageHandler) HandleBackfillTask(task *imessage.BackfillTask) {
	// TODO
}

func (imh *iMessageHandler) HandleContact(contact *imessage.Contact) {
	fmt.Printf("Received contact: %+v", contact)
}
func (imh *iMessageHandler) HandleAttachments(mtbrirdgeMsg *Message, msg *imessage.Message) {
	for _, attach := range msg.Attachments {
		fmt.Printf("Handle Attachment %s from user %s\n", attach.GetFileName(), msg.Sender.LocalID)
		data, err := attach.Read()
		if err != nil {
			imh.log.Warnln("Failed to read attachment:", err)
			continue
		}

		fileName := attach.GetFileName()
		mtbrirdgeMsg.Extra["file"] = append(mtbrirdgeMsg.Extra["file"], FileInfo{
			Name: fileName,
			Data: data,
			URL:  "",
			Size: int64(len(data)),
		})
	}
}
func (imh *iMessageHandler) Stop() {
	close(imh.stop)
}
