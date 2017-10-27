package statesync

import "github.com/gorilla/websocket"
import "fmt"
import "time"

type websocketObserver struct {
	conn   *websocket.Conn
	sendQ  chan interface{}
	closed bool
}

// NewWebsocketObserver creates an observer that sends updates over
// a websocket.
func NewWebsocketObserver(conn *websocket.Conn) Observer {
	res := &websocketObserver{
		conn:   conn,
		sendQ:  make(chan interface{}, 5),
		closed: false,
	}
	go res.sendHandler()
	return res
}

func (w *websocketObserver) sendHandler() {
	defer func() {
		w.conn.Close()
		fmt.Println("Connection closed")
	}()
	for !w.closed {
		select {
		case state := <-w.sendQ:
			if err := w.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 500)); err != nil {
				w.closed = true
				return
			}
			if err := w.conn.WriteJSON(state); err != nil {
				w.closed = true
				return
			}
			fmt.Printf("Sent %v\n", state)
		}
	}
}

func (w *websocketObserver) OnUpdate(newstate interface{}) error {
	if w.closed {
		return fmt.Errorf("Connection closed")
	}
	select {
	case w.sendQ <- newstate:
		// sent successful
	default:
		w.closed = true
		return fmt.Errorf("Connection closed")
	}
	return nil
}
