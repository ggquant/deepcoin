package deepcoin

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

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
	message   chan []byte // 订阅数据
	ctx       context.Context
	ctxCancel context.CancelFunc
	wg        sync.WaitGroup
	unpack    map[int64]UnpackFn
}

func New(opts ...Option) (*Client, error) {
	defaultOpts := options{}
	for _, o := range opts {
		o(&defaultOpts)
	}
	cCtx, cCancel := context.WithCancel(context.TODO())
	c := &Client{
		apiKey:    defaultOpts.apiKey,
		secretKey: defaultOpts.secretKey,
		baseURL:   defaultOpts.baseURL,
		id:        uuid.New(),
		message:   make(chan []byte, 256),
		ctx:       cCtx,
		ctxCancel: cCancel,
		unpack:    map[int64]UnpackFn{},
	}
	if err := c.dial(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) Send(data []byte, id int64, callback UnpackFn) error {
	c.unpack[id] = callback
	c.message <- data
	return nil
}

func (c *Client) Close() error {
	c.ctxCancel()
	c.wg.Wait()
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
		}
	}
}

func (c *Client) dial() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.baseURL, nil)
	if err != nil {
		log.Errorf("err = %v", err)
		return err
	}
	c.wg.Add(2)
	go c.read(conn)
	go c.write(conn)
	return nil
}

// 读信息，从 websocket 连接直接读取数据
func (c *Client) read(conn *websocket.Conn) {
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
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Errorf("err = %v", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if messageType == websocket.CloseMessage {
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
func (c *Client) write(conn *websocket.Conn) {
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
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	for {
		select {
		case <-c.ctx.Done():
			return
		case message, ok := <-c.message:
			if !ok {
				return
			}
			log.Infof("client [%s] write message: %s", c.id, string(message))
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Errorf("client [%s] write message err: %s", c.id, err)
			}
		}
	}
}
