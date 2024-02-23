package public

type SubscribeReq struct {
	Method string        `json:"method,omitempty"`
	Params []interface{} `json:"params,omitempty"`
	ID     int64         `json:"id,omitempty"`
}
