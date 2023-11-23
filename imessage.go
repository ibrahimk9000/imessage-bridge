package main

import (
	"log"
	"time"

	"go.mau.fi/mautrix-imessage/imessage"
)

type iMessageHandler struct {
	bridge *customBridge
	stop   chan struct{}
}

func NewiMessageHandler(bridge *customBridge) *iMessageHandler {
	return &iMessageHandler{
		bridge: bridge,
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

		select {
		case msg := <-messages:
			imh.HandleMessage(msg)
		case rr := <-readReceipts:
			imh.HandleReadReceipt(rr)
		case notif := <-typingNotifications:
			imh.HandleTypingNotification(notif)
		case chat := <-chats:
			imh.HandleChat(chat)
		case contact := <-contacts:
			imh.HandleContact(contact)
		case status := <-messageStatuses:
			imh.HandleMessageStatus(status)
		case backfillTask := <-backfillTasks:
			imh.HandleBackfillTask(backfillTask)
		case <-imh.stop:
			return
		}

	}
}

func (imh *iMessageHandler) HandleMessage(msg *imessage.Message) {
	chatID := msg.ChatGUID
	threadID := msg.ThreadID
	chat, err := imh.bridge.IM.GetChatInfo(chatID, threadID)
	if err != nil {
		log.Printf("Failed to get chat info for incoming message: %v\n", err)
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
	log.Printf("Received  message: %s , from user %s\n", msg.Text, msg.Sender.LocalID)
}

func (imh *iMessageHandler) HandleMessageStatus(status *imessage.SendMessageStatus) {
	log.Printf("Received message status: %+v", status)
}

func (imh *iMessageHandler) HandleReadReceipt(rr *imessage.ReadReceipt) {
	log.Printf("Received read receipt: %+v", rr)
}

func (imh *iMessageHandler) HandleTypingNotification(notif *imessage.TypingNotification) {
	log.Printf("Received typing notification: %+v", notif)
}

func (imh *iMessageHandler) HandleChat(chat *imessage.ChatInfo) {
	log.Printf("Received chat info: %+v", chat)
}

func (imh *iMessageHandler) HandleBackfillTask(task *imessage.BackfillTask) {
	// TODO
}

func (imh *iMessageHandler) HandleContact(contact *imessage.Contact) {
	log.Printf("Received contact: %+v", contact)
}
func (imh *iMessageHandler) HandleAttachments(mtbrirdgeMsg *Message, msg *imessage.Message) {
	for _, attach := range msg.Attachments {
		log.Printf("Handle Attachment %s from user %s\n", attach.GetFileName(), msg.Sender.LocalID)
		data, err := attach.Read()
		if err != nil {
			log.Printf("Failed to read attachment:%s", err)
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
