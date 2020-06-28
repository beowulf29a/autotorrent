package main

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

////////// input
/*
   listen on MQTT exchange for
	- magnet links
	- stop
*/
////////// output
/*
	publish updates on MQTT exchange for
	 - Download updates
	   - date / time started
	   - current bytes
	   - total bytes
	   - current download rate
*/
const (
	AddTorrent  uint8 = 1
	StopTorrent uint8 = 2
)

const (
	topicSub string = "autotorrent/subscriber"
	topicPub string = "autotorrent/receiver"
	broker   string = "192.168.86.50:1883"
	user     string = "mqttuser"
	pwd      string = "bojangles"
	id       string = "autotorrent"
)

func main() {
	client := createMQTTClient()
	sub := NewSub(client)
	pub := NewPub(client)
	pub.init()

	//main loop for subscriber
	//listen for new requests or stop requests
	for {
		//block for new message
		msg := <-sub.MessageAlert

		fmt.Print(msg.Type)
		fmt.Print(msg.Name)
		if msg.Type == AddTorrent {
			go pub.AddTorrent(msg.Name)
		} else if msg.Type == StopTorrent {
			go pub.RemoveTorrent(msg.Name)
		}
	}
}

func createMQTTClient() mqtt.Client {
	opts := &mqtt.ClientOptions{}

	opts.AddBroker(broker)
	opts.SetClientID(id)
	opts.SetUsername(user)
	opts.SetPassword(pwd)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return client
}
