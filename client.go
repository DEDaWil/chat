package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	id   uuid.UUID       // для идентификации (может пригодиться в будущем)
	hub  *Hub            // ссылка на “центр чата”
	conn *websocket.Conn // настоящее WebSocket соединение
	send chan []byte     // канал, по которому сервер будет слать сообщения клиенту
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		id:   uuid.New(),
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

// клиент ПРИНИМАЕТ сообщения
//
// Этот цикл:
//
//	ждёт, когда клиент отправит что-нибудь
//	читает сообщение
//	передаёт его хабу для рассылки
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		c.hub.broadcast <- message
	}
}

// клиент ОТПРАВЛЯЕТ сообщения
//
// Этот цикл:
//
//	слушает канал send
//	всё, что туда прилетает, пишется в WebSocket клиенту
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// канал закрыт
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
