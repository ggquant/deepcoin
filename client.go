package deepcoin

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/go-bamboo/pkg/log"
	"github.com/go-bamboo/pkg/uuid"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type UnpackFn func(data []byte) error

// Client 单个 websocket 信息
type Client struct {
	apiKey    string
	secretKey string
	baseURL   string
	id        string
	rpcId     int64
	conn      *websocket.Conn
	lock      sync.RWMutex
	message   chan []byte // 订阅数据
	feed      event.Feed
	ctx       context.Context
	ctxCancel context.CancelFunc
	wg        sync.WaitGroup
	unpack    map[int64]UnpackFn
}

func New(opts ...Option) *Client {
	defaultOpts := options{}
	for _, o := range opts {
		o(&defaultOpts)
	}
	cCtx, cCancel := context.WithCancel(context.TODO())
	return &Client{
		apiKey:    defaultOpts.apiKey,
		secretKey: defaultOpts.secretKey,
		baseURL:   defaultOpts.baseURL,
		id:        uuid.New(),
		message:   make(chan []byte, 256),
		ctx:       cCtx,
		ctxCancel: cCancel,
		unpack:    map[int64]UnpackFn{},
	}
}

func (c *Client) Send(data []byte, id int64, callback UnpackFn) error {
	c.unpack[id] = callback
	c.message <- data
	return nil
}

func (c *Client) Subscribe(ch interface{}) event.Subscription {
	return c.feed.Subscribe(ch)
}

func (c *Client) Start() error {
	c.wg.Add(3)
	go c.watchConn()
	go c.read()
	go c.write()
	return nil
}

func (c *Client) Stop() error {
	c.ctxCancel()
	c.wg.Wait()
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Errorf("client [%s] disconnect err: %s", c.id, err)
			return err
		}
	}
	return nil
}

func (c *Client) watchConn() {
	defer func() {
		c.wg.Done()
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			pl := fmt.Sprintf("client watchConn call panic: %v\n%s\n", err, buf)
			log.Errorf("%s", pl)
		}
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// try conn
			c.lock.Lock()
			if c.conn == nil {
				conn, _, err := websocket.DefaultDialer.Dial(c.baseURL, nil)
				if err != nil {
					log.Errorf("err = %v", err)
					c.lock.Unlock()
					time.Sleep(1 * time.Second)
					continue
				} else {
					c.conn = conn
					c.lock.Unlock()
				}
			} else {
				c.lock.Unlock()
			}

		}
	}
}

// 读信息，从 websocket 连接直接读取数据
func (c *Client) read() {
	defer func() {
		c.wg.Done()
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			pl := fmt.Sprintf("client read call panic: %v\n%s\n", err, buf)
			log.Errorf("%s", pl)
		}
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.lock.RLock()
			if c.conn == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			messageType, message, err := c.conn.ReadMessage()
			c.lock.RUnlock()
			if err != nil {
				log.Errorf("err = %v", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if messageType == websocket.CloseMessage {
				c.conn = nil
				time.Sleep(1 * time.Second)
				continue
			}
			// c.log.Infof("client [%s] receive message: %s", c.Id, string(message))
			localNo := gjson.GetBytes(message, "localNo")
			fn, ok := c.unpack[localNo.Int()]
			if ok {
				if err := fn(message); err != nil {
					log.Error(err)
				}
			}
		}
	}
}

// 写信息，从 channel 变量 Send 中读取数据写入 websocket 连接
func (c *Client) write() {
	defer func() {
		c.wg.Done()
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			pl := fmt.Sprintf("client write call panic: %v\n%s\n", err, buf)
			log.Errorf("%s", pl)
		}
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case message, ok := <-c.message:
			if !ok {
				c.lock.RLock()
				if c.conn == nil {
					c.lock.RUnlock()
					return
				}
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Errorf("err = %v", err)
				}
				c.lock.RUnlock()
				return
			}
			log.Infof("client [%s] write message: %s", c.id, string(message))
			c.lock.RLock()
			if c.conn == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			c.lock.RUnlock()
			if err != nil {
				log.Errorf("client [%s] write message err: %s", c.id, err)
			}
		}
	}
}
