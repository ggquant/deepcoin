package deepcoin

import (
	"github.com/go-bamboo/pkg/log"
	_ "github.com/go-bamboo/pkg/log/std"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewStream(t *testing.T) {
	s := NewStream("", "")
	c, err := s.Public().Swap().Client()
	assert.NoError(t, err)

	c.Kline(func(data *PushKLine) error {
		log.Infof("%v", data.Action)
		return nil
	})

	for {
		time.Sleep(1 * time.Second)
	}
}
