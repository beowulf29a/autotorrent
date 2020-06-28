package main

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/anacrolix/torrent"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//MQTTPub -
type MQTTPub struct {
	mqttClient    mqtt.Client
	torrentClient *torrent.Client
	mapLocker     sync.Mutex
	torrents      map[string]AutoTorrent
	updateChan    chan TMessage
}

//NewPub -
func NewPub(client mqtt.Client) MQTTPub {
	return MQTTPub{
		mqttClient: client,
		torrents:   make(map[string]AutoTorrent),
		updateChan: make(chan TMessage),
	}
}

func (p *MQTTPub) init() {
	config := torrent.NewDefaultClientConfig()
	config.Seed = true
	c, err := torrent.NewClient(config)
	if nil == err {
		p.torrentClient = c
		go p.PublishToMQTT()
	}

}

//AddTorrent -
func (p *MQTTPub) AddTorrent(link string) {
	guid := p.getGUID(link)
	if 0 < len(link) {
		t, e := p.torrentClient.AddMagnet(link)
		if nil == e {
			if _, ok := p.torrents[guid]; !ok {
				p.mapLocker.Lock()
				p.torrents[guid] = NewAutoTorrent(guid, t, p.updateChan)
				p.mapLocker.Unlock()
				p.torrents[guid].StartTorrent()
			}
		}
	}
}

func (p *MQTTPub) getGUID(val string) string {
	idx := strings.LastIndex(val, ":") + 1
	return val[idx:]
}

//PublishToMQTT -
func (p *MQTTPub) PublishToMQTT() {
	updateMap := make(map[string]TMessage)

	var err error

	for {
		msg, ok := <-p.updateChan
		if !ok {
			break
		} else {
			updateMap[msg.GUID] = msg

			//synchronize maps
			if len(p.torrents) != len(updateMap) {
				for k := range p.torrents {
					var found = false
					for u := range updateMap {
						if k == u {
							found = true
							break
						}
					}
					if !found {
						delete(updateMap, k)
					}
				}
			}

			if nil == err {
				var output = make([]TMessage, len(updateMap))
				var count int8 = 0
				for _, v := range updateMap {
					output[count] = v
					count++
				}
				b, _ := json.Marshal(fml{Entities: output})
				p.mqttClient.Publish(topicPub, 0, false, b)
			}
		}
	}
}

type fml struct {
	Entities []TMessage
}

//RemoveTorrent -
func (p *MQTTPub) RemoveTorrent(guid string) {
	p.mapLocker.Lock()
	if t, ok := p.torrents[guid]; ok {
		t.StopTorrent()
		t = nil
		delete(p.torrents, guid)
	}
	p.mapLocker.Unlock()
}
