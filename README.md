# IMessage  bridge
The imessage bridge is part of bridge setup , it  need to be excuted on the macos with imessage logged in .it exchanges messages between the imessage and remote matterbridge server. 
# Setup 
Note : the imessage on the macos should be logged in

## Build  the bridge on macos
```bash
git clone https://github.com/ibrahimk9000/imessage-bridge.git
cd imessage-bridge
go mod tidy
go build
```
## Config the bridge disk access
* add the binary to full disk access list
  
![full disk access](https://www.bitdefender.com/media/uploads/2018/08/plus-sign-full-disk-access.png)

## Create a matterbridge imessage network

* Create imessage bridge from the  guide here
  * [create new matterbridge network](https://github.com/pingponglabs/dendrite-endpoint/blob/doc-update/register-guide.MD##create-matterbridge-network-and-joined-it-with-matrix-bot-room) 
  * [imessage network json config example](https://github.com/pingponglabs/dendrite-endpoint/blob/doc-update/register-guide.MD#imessage)
* Get the imessage command from the status endpoint and execute it on the macOS terminal

```text
example : ./imessage-bridge -account ibra1990ski@gmail.com -remote-api-address https://matrix.mediamagic.ai
```
