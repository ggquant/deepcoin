package deepcoin

import (
	"encoding/json"
	"sync/atomic"
)

type PushKLine struct {
	Action string
}

func (c *Client) Kline(cb func(data *PushKLine) error) error {
	req := SendTopicAction{
		Action: "1",
	}
	rpcID := atomic.AddInt64(&c.rpcId, 1)
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	fn, err := c.unpackKline(cb)
	if err != nil {
		return err
	}
	return c.Send(data, rpcID, fn)
}

func (c *Client) unpackKline(cb func(data *PushKLine) error) (UnpackFn, error) {
	return func(data []byte) error {
		var ret PushKLine
		if err := json.Unmarshal(data, &ret); err != nil {
			return err
		}
		return cb(&ret)
	}, nil
}
