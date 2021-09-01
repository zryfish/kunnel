package utils

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/klog"
)

type wsConn struct {
	*websocket.Conn
	buff []byte
}

func NewWebSocketConn(websocketConn *websocket.Conn) net.Conn {
	c := wsConn{
		Conn: websocketConn,
	}
	return &c
}

func (c *wsConn) Read(dst []byte) (int, error) {
	ldst := len(dst)
	var src []byte
	if l := len(c.buff); l > 0 {
		src = c.buff
		c.buff = nil
	} else {
		t, msg, err := c.ReadMessage()
		if err != nil {
			return 0, err
		} else if t != websocket.BinaryMessage {
			klog.Warning(" recieved non-binary message")
		} else {
			src = msg
		}
	}

	var copied int
	if len(src) > ldst {
		copied = copy(dst, src[:ldst])
		remain := src[ldst:]
		c.buff = make([]byte, len(remain))
		copy(c.buff, remain)
	} else {
		copied = copy(dst, src)
	}
	return copied, nil
}

func (c *wsConn) Write(b []byte) (int, error) {
	if err := c.Conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *wsConn) SetDeadline(t time.Time) error {
	if err := c.Conn.SetReadDeadline(t); err != nil {
		return err
	}
	return nil
}
