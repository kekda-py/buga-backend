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
		case m := <-broadcast:
			for c := range connections {
        log.Println("checking: ", m)
        if c == m.by {
          continue
        }
        log.Println("Sending: ", m.content)
				if err := c.WriteMessage(websocket.TextMessage, []byte(m.content)); err != nil {
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
			// MakeMessage(string(m), c)
      broadcast <- message{
        content : string(m),
        by : c,
      }
		} else {
			log.Println("websocket message received of type", mt)
		}
	}
}
