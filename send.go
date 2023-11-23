package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.mau.fi/mautrix-imessage/imessage"
)

type Message struct {
	Text      string    `json:"text"`
	Channel   string    `json:"channel"`
	Username  string    `json:"username"`
	UserID    string    `json:"userid"` // userid on the bridge
	Avatar    string    `json:"avatar"`
	Account   string    `json:"account"`
	Event     string    `json:"event"`
	Protocol  string    `json:"protocol"`
	Gateway   string    `json:"gateway"`
	ParentID  string    `json:"parent_id"`
	Timestamp time.Time `json:"timestamp"`
	ID        string    `json:"id"`
	Extra     map[string][]interface{}

	ExtraNetworkInfo
}
type FileInfo struct {
	Name     string
	Data     []byte
	Comment  string
	URL      string
	Size     int64
	Avatar   bool
	SHA      string
	NativeID string
}

type ExtraNetworkInfo struct {
	ChannelUsersMember []string          `json:"channel_users_member,omitempty"`
	ActionCommand      string            `json:"action_command,omitempty"`
	ChannelId          string            `json:"channel_id,omitempty"`
	ChannelName        string            `json:"channel_name,omitempty"`
	ChannelType        string            `json:"channel_type,omitempty"`
	TargetPlatform     string            `json:"target_platform,omitempty"`
	UsersMemberId      map[string]string `json:"users_member_id,omitempty"`
	Mentions           map[string]string `json:"mentions,omitempty"`
}

// TODO implement long polling to matterbridge api
// https://imessage.matrix.mediamagic.ai/api/message
var (
	account          string
	remoteApiAddress string
)

func init() {
	flag.StringVar(&account, "account", "", "remoteApi account to use")
	flag.StringVar(&remoteApiAddress, "remote-api-address", "", "Address to listen for remote API connections on")
}
func (c *customBridge) HandleRemoteApiSend(msg Message) {

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(msg)
	if err != nil {
		log.Printf("Error encoding message : %v\n", err)
		return
	}

	resp, err := http.Post(remoteApiAddress+"/webhook/imessage/message/"+c.remoteApiAccountB64, "application/json", &buf)
	if err != nil {
		log.Printf("Error sending message to remote bridge : %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Println(string(body))
}

func (c *customBridge) HandleRemoteApiReceive() {
	client := &http.Client{Timeout: 60 * time.Second}
	log.Println("ready to receive requests")
	for {
		respGet, err := client.Get(remoteApiAddress + "/webhook/imessage/custom-stream/" + c.remoteApiAccountB64)
		if err != nil {
			log.Printf("Error getting response  from remote bridge : %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer respGet.Body.Close()

		bodyGet, err := io.ReadAll(respGet.Body)
		if err != nil {
			log.Printf("Error reading body: %v\n", err)

		}
		msg := Message{}
		err = json.Unmarshal(bodyGet, &msg)
		if err != nil {
			log.Printf("Error unmarshalling message: %v\n", err)

		}

		c.HandleSend(msg)

		if c.Stop {
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func (c *customBridge) HandleSend(msg Message) {
	if msg.Extra["file"] != nil {
		c.HandleFileSend(msg)
		return
	}
	if msg.Text != "" {
		fmt.Printf("Received  message: %s , from user %s\n", msg.Text, msg.Channel)
		c.HandleTextSend(msg)
	}

}

func (c *customBridge) HandleFileSend(msg Message) {
	for _, f := range msg.Extra["file"] {
		fb, err := json.Marshal(f)
		if err != nil {
			log.Println(err)
			return
		}
		fi := FileInfo{}
		err = json.Unmarshal(fb, &fi)
		if err != nil {
			log.Printf("Error unmarshalling file: %v\n", err)
			return
		}

		mtype := http.DetectContentType(fi.Data[:512])
		path := filepath.Join(os.TempDir(), fi.Name)
		err = os.WriteFile(path, fi.Data, 0644)
		if err != nil {
			log.Printf("Error writing file: %v\n", err)
			return
		}
		fmt.Printf("Handle Attachment %s from user %s\n", fi.Name, msg.Channel)

		_, err = c.IM.SendFile(msg.Channel, "", fi.Name, path, "", 0, mtype, false, imessage.MessageMetadata{})
		if err != nil {
			log.Printf("Error sending file: %v\n", err)
			return
		}

	}
}
func (c *customBridge) HandleTextSend(msg Message) {
	c.IM.SendMessage(msg.Channel, msg.Text, "", 0, nil, nil)

}

func (c *customBridge) HandleRemoteApiListen() {

}

func (c *customBridge) SendChatsInfoToRemote(chatInfo *imessage.ChatInfo) {

	channelID := chatInfo.Identifier.String()
	channelName := chatInfo.DisplayName
	if channelName == "" {
		channelName = chatInfo.LocalID
	}

	userMembersID := map[string]string{}
	event := ""
	for _, member := range chatInfo.Members {
		contact, _ := c.IM.GetContactInfo(member)
		if contact != nil {
			username := contact.FirstName + " " + contact.LastName
			userMembersID[member] = username
		} else {
			userMembersID[member] = member
		}
	}
	username := ""
	userID := ""
	if chatInfo.IsGroup {

		event = "new_users"
	} else {
		event = "direct_msg"
		for k, v := range userMembersID {
			username = v
			userID = k
			break
		}

	}
	mtbridgeMsg := Message{
		Text:      "new_users",
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
		ID:        chatInfo.Identifier.String(),
		Extra:     map[string][]interface{}{},
		ExtraNetworkInfo: ExtraNetworkInfo{
			ActionCommand:  "imessage",
			ChannelId:      channelID,
			ChannelName:    channelName,
			TargetPlatform: "appservice",

			UsersMemberId: userMembersID,
		},
	}
	c.HandleRemoteApiSend(mtbridgeMsg)
}
