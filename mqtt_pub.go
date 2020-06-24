package main

import (
	"sync"

	"github.com/anacrolix/torrent"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTPub struct {
	mqttClient    mqtt.Client
	torrentClient *torrent.Client
	mapLocker     sync.Mutex
	torrents      map[string]AutoTorrent
}

func NewPub(client mqtt.Client) MQTTPub {
	return MQTTPub{
		mqttClient: client,
	}
}

func (p *MQTTPub) init() {
	config := torrent.NewDefaultClientConfig()
	config.Seed = true
	p.torrentClient, _ = torrent.NewClient(config)
}

func (p *MQTTPub) AddTorrent(link string, guid string) {
	if 0 < len(link) {
		t, e := p.torrentClient.AddMagnet(link)
		if nil == e {
			if _, ok := p.torrents[link]; !ok {
				p.mapLocker.Lock()
				p.torrents[guid] = NewAutoTorrent(guid, t, p.mqttClient)
				p.mapLocker.Unlock()
				p.torrents[guid].StartTorrent()
			}
		}
	}
}

func (p *MQTTPub) RemoveTorrent(guid string) {
	if t, ok := p.torrents[guid]; ok {
		t.StopTorrent()
	}
}
