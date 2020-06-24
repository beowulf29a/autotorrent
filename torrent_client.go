package main

import (
	"encoding/json"
	"time"

	"github.com/anacrolix/torrent"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type AutoTorrent interface {
	//guid byte array
	StopTorrent()
	StartTorrent()
}

//autotorrent implements the AutoTorrent interface
type atorrent struct {
	Guid      string
	MQClient  mqtt.Client
	Torrent   *torrent.Torrent
	StartTime int64
	close     bool
}

type TMessage struct {
	Name           string  `json:"n"`
	Guid           string  `json:"g"`
	BytesTotal     int64   `json:"bt"`
	BytesComplete  int64   `json:"bc"`
	DownloadSpeed  float32 `json:"down"`
	ConnectedPeers int     `json:"peers"`
	StartTime      int64   `json:"t"`
}

func NewAutoTorrent(guid string, torrent *torrent.Torrent, mc mqtt.Client) AutoTorrent {
	at := &atorrent{
		Guid:      guid,
		MQClient:  mc,
		Torrent:   torrent,
		StartTime: time.Now().Unix(),
	}
	return at
}

func (at *atorrent) StopTorrent() {
	at.Torrent.Drop()
	at.close = true
}

func (at *atorrent) StartTorrent() {
	<-at.Torrent.GotInfo()
	at.Torrent.DownloadAll()

	var (
		lastDOwnload int64 = 0
		totalSize    int64 = at.Torrent.Info().TotalLength()
	)
	for at.Torrent.BytesMissing() != 0 && !at.close {
		outmsg, _ := json.Marshal(TMessage{
			Name:           at.Torrent.Name(),
			Guid:           at.Guid,
			BytesTotal:     totalSize,
			BytesComplete:  at.Torrent.BytesCompleted(),
			DownloadSpeed:  float32(at.Torrent.BytesCompleted()-lastDOwnload) / 2000.0,
			ConnectedPeers: at.Torrent.Stats().TotalPeers,
			StartTime:      at.StartTime,
		})
		at.MQClient.Publish(topic_pub, 0, false, outmsg)
		time.Sleep(2 * time.Second)
		lastDOwnload = at.Torrent.BytesCompleted()
	}

	//do we stop or let it run?
	//at.StopTorrent()
}
