package deepcoin

import "net/url"

type Stream struct {
	apiKey    string
	secretKey string
	url       string
}

func NewStream(apiKey, secretKey string) *Stream {
	url := "wss://stream.deepcoin.com"
	return &Stream{
		apiKey:    apiKey,
		secretKey: secretKey,
		url:       url,
	}
}

func (s *Stream) Public() *Stream {
	u, _ := url.JoinPath(s.url, "public")
	return &Stream{
		apiKey:    s.apiKey,
		secretKey: s.secretKey,
		url:       u,
	}
}

func (s *Stream) Swap() *Stream {
	u, _ := url.JoinPath(s.url, "ws")
	return &Stream{
		apiKey:    s.apiKey,
		secretKey: s.secretKey,
		url:       u,
	}
}

func (s *Stream) Spot() *Stream {
	u, _ := url.JoinPath(s.url, "spotws")
	return &Stream{
		apiKey:    s.apiKey,
		secretKey: s.secretKey,
		url:       u,
	}
}

func (s *Stream) Client() *Client {
	return &Client{
		apiKey:    s.apiKey,
		secretKey: s.secretKey,
		baseURL:   s.url,
	}
}
