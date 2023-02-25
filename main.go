package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"time"

	"go.mau.fi/mautrix-imessage/imessage"
	_ "go.mau.fi/mautrix-imessage/imessage/mac"

	"go.mau.fi/mautrix-imessage/ipc"
	log "maunium.net/go/maulogger/v2"
)

func main() {

	flag.Parse()

	var im = &customBridge{
		Log: log.Create(),
	}
	im.remoteApiAccount = account
	im.remoteApiAddress = remoteApiAddress
	im.remoteApiAccountB64 = base64.URLEncoding.EncodeToString([]byte(account))

	br, err := imessage.NewAPI(im)
	if err != nil {
		fmt.Println(err)
	}

	im.IM = br

	im.MsgHandler = NewiMessageHandler(im)

	tt, err := im.IM.GetChatsWithMessagesAfter(time.Now().Add(-time.Hour * 240 * 4 * 7))
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range tt {
		fmt.Println(v.ChatGUID)
		fmt.Println(v.ThreadID)
		ch, err := im.IM.GetChatInfo(v.ChatGUID, v.ThreadID)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(ch.Members)
		im.SendChatsInfoToRemote(ch)

	}

	go func() {
		err = im.IM.Start(func() {
			fmt.Println("imessage started")
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		time.Sleep(time.Second * 10)
	}()
	go im.HandleRemoteApiReceive()
	im.MsgHandler.Start()

}

type customBridge struct {
	IM                  imessage.API
	MsgHandler          *iMessageHandler
	Log                 log.Logger
	remoteApiAccount    string
	remoteApiAccountB64 string
	remoteApiAddress    string
	Stop                bool
}

var _ imessage.Bridge = (*customBridge)(nil)

func (c *customBridge) GetIPC() *ipc.Processor {
	return nil
}
func (c *customBridge) GetLog() log.Logger {
	// TODO: Implement
	return c.Log
}
func (c *customBridge) GetConnectorConfig() *imessage.PlatformConfig {
	return &imessage.PlatformConfig{
		Platform:     "mac",
		ContactsMode: "mac",
	}
}
func (c *customBridge) PingServer() (start, serverTs, end time.Time) {
	return
}
func (c *customBridge) SendBridgeStatus(state imessage.BridgeStatus) {
	return
}
func (c *customBridge) ReIDPortal(oldGUID, newGUID string, mergeExisting bool) bool {
	return false
}
func (c *customBridge) GetMessagesSince(chatGUID string, since time.Time) []string {
	// TODO: Implement
	return nil
}
func (c *customBridge) SetPushKey(req *imessage.PushKeyRequest) {
	return
}
