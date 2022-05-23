package main

import (
	"log"

	"github.com/gofiber/websocket/v2"
)

func WebsocketHub() {
	for {
		select {
		case c := <-register:
			connections[c] = client{}
			log.Println("client registered")
		case msg := <-broadcast:
			for c := range connections {
				if c == msg.by {
					continue
				}
				if err := c.WriteMessage(websocket.TextMessage, []byte(msg.content)); err != nil {
					log.Println("Error while sending message: ", err)

					c.WriteMessage(websocket.CloseMessage, []byte{})
					c.Close()
					delete(connections, c)
				}
			}
		case c := <-unregister:
			delete(connections, c)

			log.Println("client unregistered")
		}
	}
}

func WebSocket(c *websocket.Conn) {
	defer func() {
		unregister <- c
		c.Close()
	}()

	register <- c

	for {
		mt, m, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}

			return // Calls the deferred function, i.e. closes the connection on error
		}

		if mt == websocket.TextMessage {
			MakeMessage(string(m), c)
		} else {
			log.Println("websocket message received of type", mt)
		}
	}
}
