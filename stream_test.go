package deepcoin

import (
	"github.com/go-bamboo/pkg/log"
	_ "github.com/go-bamboo/pkg/log/std"
	"testing"
	"time"
)

func TestNewStream(t *testing.T) {
	s := NewStream("", "")
	c := s.Public().Swap().Client()
	c.Start()

	c.Kline(func(data *PushKLine) error {
		log.Infof("%v", data.Action)
		return nil
	})

	for {
		time.Sleep(1 * time.Second)
	}
}
