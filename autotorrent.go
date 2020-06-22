package main

import (
	"fmt"
	"log"
	"time"

	"github.com/anacrolix/torrent"
)

func main() {
	c, _ := torrent.NewClient(nil)
	defer c.Close()
	t, _ := c.AddMagnet("magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU")
	<-t.GotInfo()
	t.DownloadAll()
	go func() {
		var lastDOwnload int64 = 0
		for t.BytesMissing() > 10 {
			var speed = lastDOwnload / 2
			var pct = float64(t.BytesCompleted()) / float64(t.Info().TotalLength())
			var remainingTime = float64(t.Info().TotalLength()) / float64(speed)
			fmt.Printf("%2.2f%% %v - %v :: %vkbps :: %2.2fsecons remain\n", 100*pct, t.BytesMissing(), t.Info().TotalLength(), speed/1000, remainingTime)
			time.Sleep(2 * time.Second)
			lastDOwnload = t.BytesCompleted() - lastDOwnload
		}
	}()
	c.WaitAll()
	log.Print("ermahgerd, torrent downloaded")
}
