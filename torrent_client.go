package main

import (
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
}

type TMessage struct {
	BytesTotal     int64   `json:"bt"`
	BytesComplete  int64   `json:"bc"`
	DownloadSpeed  float32 `json:"down"`
	ConnectedPeers int     `json:"peers"`
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
}

func (at *atorrent) StartTorrent() {
	<-at.Torrent.GotInfo()
	at.Torrent.DownloadAll()

	var (
		lastDOwnload int64 = 0
		totalSize    int64 = at.Torrent.Info().TotalLength()
	)
	for at.Torrent.BytesMissing() == 0 {
		at.MQClient.Publish(topic_pub, 0, false, TMessage{
			BytesTotal:     totalSize,
			BytesComplete:  at.Torrent.BytesCompleted(),
			DownloadSpeed:  float32(lastDOwnload) / 2000.0,
			ConnectedPeers: at.Torrent.Stats().TotalPeers,
		})
		time.Sleep(2 * time.Second)
		lastDOwnload = at.Torrent.BytesCompleted()
	}

	//do we stop or let it run?
	//at.StopTorrent()
}
