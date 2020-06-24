package main

import (
	"encoding/json"
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var pMsg SubMessage
	json.Unmarshal(msg.Payload(), &pMsg)
	pubChan <- pMsg
}

var (
	pubChan        chan SubMessage
	messageRecChan chan string
)

type SubMessage struct {
	Type uint8  `json:"type"`
	Name string `json:"name"` //can be guid or magnet link
}

type MQTTSub struct {
	MessageAlert chan SubMessage
	client       mqtt.Client
}

//returns a channel which will fire on
// - 0: new torrent
// - 1: stop torrent
func NewSub(client mqtt.Client) MQTTSub {
	pubChan = make(chan SubMessage)

	if token := client.Subscribe(topic_sub, 0, f); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	return MQTTSub{
		MessageAlert: pubChan,
		client:       client,
	}

}

func (mqttsub *MQTTSub) ShutDown() {
	mqttsub.client.Disconnect(0)
}
