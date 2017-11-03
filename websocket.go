package statesync

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type websocketObserver struct {
	conn   *websocket.Conn
	sendQ  chan interface{}
	recvQ  chan interface{}
	closed bool
}

// NewWebsocketObserver creates an observer that sends updates over
// a websocket.
func NewWebsocketObserver(conn *websocket.Conn) Observer {
	res := &websocketObserver{
		conn:   conn,
		sendQ:  make(chan interface{}, 5),
		recvQ:  make(chan interface{}, 5),
		closed: false,
	}
	go res.sendHandler()
	go res.recvHandler()
	return res
}

func (w *websocketObserver) sendHandler() {
	defer func() {
		w.conn.Close()
		close(w.recvQ)
	}()
	for !w.closed {
		select {
		case state, ok := <-w.sendQ:
			if !ok {
				w.closed = true
				return
			}
			if err := w.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 500)); err != nil {
				w.closed = true
				return
			}
			if err := w.conn.WriteJSON(state); err != nil {
				w.closed = true
				return
			}
		}
	}
}

func (w *websocketObserver) recvHandler() {
	for !w.closed {
		var v interface{}
		if err := w.conn.ReadJSON(&v); err != nil {
			w.closed = true
			return
		}
		w.recvQ <- v
	}
}

func (w *websocketObserver) OnUpdate(newstate interface{}) error {
	select {
	case w.sendQ <- newstate:
		// sent successful
	default:
		w.closed = true
		close(w.sendQ)
		return fmt.Errorf("Connection closed")
	}
	return nil
}

func (w *websocketObserver) ReceiveChan() <-chan interface{} {
	return w.recvQ
}
