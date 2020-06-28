package main

import (
	"time"

	"github.com/anacrolix/torrent"
)

//AutoTorrent -
type AutoTorrent interface {
	//guid byte array
	StopTorrent()
	StartTorrent()
}

//autotorrent implements the AutoTorrent interface
type atorrent struct {
	GUID       string
	UpdateChan chan TMessage
	Torrent    *torrent.Torrent
	StartTime  int64
	close      bool
}

//TMessage -
type TMessage struct {
	Name           string  `json:"name"`
	GUID           string  `json:"guid"`
	BytesTotal     int64   `json:"bytetot"`
	BytesComplete  int64   `json:"bytecom"`
	DownloadSpeed  float32 `json:"down"`
	Percent        float32 `json:"pct"`
	ConnectedPeers int     `json:"peers"`
	StartTime      int64   `json:"st"`
	LastUpdate     int64   `json:"lastupdate"`
}

//NewAutoTorrent -
func NewAutoTorrent(guid string, torrent *torrent.Torrent, updateCh chan TMessage) AutoTorrent {
	at := &atorrent{
		GUID:       guid,
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
			GUID:           at.GUID,
			BytesTotal:     totalSize,
			BytesComplete:  at.Torrent.BytesCompleted(),
			DownloadSpeed:  float32(at.Torrent.BytesCompleted()-lastDOwnload) / 2000.0,
			ConnectedPeers: at.Torrent.Stats().TotalPeers,
			StartTime:      at.StartTime,
			LastUpdate:     time.Now().Unix(),
			Percent:        100.0 * float32(at.Torrent.BytesCompleted()) / float32(totalSize),
		}

		time.Sleep(2 * time.Second)
		lastDOwnload = at.Torrent.BytesCompleted()
	}

	//should be 100%
	at.UpdateChan <- TMessage{
		Name:           at.Torrent.Name(),
		GUID:           at.GUID,
		BytesTotal:     totalSize,
		BytesComplete:  at.Torrent.BytesCompleted(),
		DownloadSpeed:  float32(at.Torrent.BytesCompleted()-lastDOwnload) / 2000.0,
		ConnectedPeers: at.Torrent.Stats().TotalPeers,
		StartTime:      at.StartTime,
		LastUpdate:     time.Now().Unix(),
		Percent:        100.0 * float32(at.Torrent.BytesCompleted()) / float32(totalSize),
	}

	//do we stop or let it run?
	//at.StopTorrent()
}
