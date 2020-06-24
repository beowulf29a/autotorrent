package main

import (
	"time"

	"github.com/anacrolix/torrent"
)

type AutoTorrent interface {
	//guid byte array
	StopTorrent()
	StartTorrent()
}

//autotorrent implements the AutoTorrent interface
type atorrent struct {
	Guid       string
	UpdateChan chan TMessage
	Torrent    *torrent.Torrent
	StartTime  int64
	close      bool
}

type TMessage struct {
	Name           string  `json:"n"`
	Guid           string  `json:"g"`
	BytesTotal     int64   `json:"bt"`
	BytesComplete  int64   `json:"bc"`
	DownloadSpeed  float32 `json:"down"`
	ConnectedPeers int     `json:"peers"`
	StartTime      int64   `json:"t"`
	LastUpdate     int64   `json:"l"`
}

func NewAutoTorrent(guid string, torrent *torrent.Torrent, updateCh chan TMessage) AutoTorrent {
	at := &atorrent{
		Guid:       guid,
		UpdateChan: updateCh,
		Torrent:    torrent,
		StartTime:  time.Now().Unix(),
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
		//inform update chan that new info is ready
		at.UpdateChan <- TMessage{
			Name:           at.Torrent.Name(),
			Guid:           at.Guid,
			BytesTotal:     totalSize,
			BytesComplete:  at.Torrent.BytesCompleted(),
			DownloadSpeed:  float32(at.Torrent.BytesCompleted()-lastDOwnload) / 2000.0,
			ConnectedPeers: at.Torrent.Stats().TotalPeers,
			StartTime:      at.StartTime,
			LastUpdate:     time.Now().Unix(),
		}

		time.Sleep(2 * time.Second)
		lastDOwnload = at.Torrent.BytesCompleted()
	}

	//do we stop or let it run?
	//at.StopTorrent()
}
