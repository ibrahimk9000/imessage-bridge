package main

import (
	"encoding/base64"
	"flag"
	"time"

	"go.mau.fi/mautrix-imessage/imessage"
	_ "go.mau.fi/mautrix-imessage/imessage/mac"

	"log"

	"go.mau.fi/mautrix-imessage/ipc"
	mlog "maunium.net/go/maulogger/v2"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	var im = &customBridge{
		Log: mlog.Create(),
	}
	im.remoteApiAccount = account
	im.remoteApiAddress = remoteApiAddress
	im.remoteApiAccountB64 = base64.URLEncoding.EncodeToString([]byte(account))

	br, err := imessage.NewAPI(im)
	if err != nil {
		log.Fatal(err)
	}

	im.IM = br

	im.MsgHandler = NewiMessageHandler(im)

	tt, err := im.IM.GetChatsWithMessagesAfter(time.Now().Add(-time.Hour * 240 * 4 * 7))
	if err != nil {
		log.Println(err)
	}
	log.Println("list of available chats")
	for _, v := range tt {
		log.Printf("chat ID :%s", v.ChatGUID)
		log.Printf("thread ID :%s", v.ThreadID)

		ch, err := im.IM.GetChatInfo(v.ChatGUID, v.ThreadID)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("chat members :%v", ch.Members)
		log.Println("-------------------------------------------------------------------------------")

	}

	go func() {
		err = im.IM.Start(func() {
			log.Println("imessage bridge  started")
		})
		if err != nil {
			log.Fatal(err)
		}
	}()
	go im.HandleRemoteApiReceive()
	log.Println("starting message handler")
	im.MsgHandler.Start()

}

type customBridge struct {
	IM                  imessage.API
	MsgHandler          *iMessageHandler
	Log                 mlog.Logger
	remoteApiAccount    string
	remoteApiAccountB64 string
	remoteApiAddress    string
	Stop                bool
}

var _ imessage.Bridge = (*customBridge)(nil)

func (c *customBridge) GetIPC() *ipc.Processor {
	return nil
}
func (c *customBridge) GetLog() mlog.Logger {
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
