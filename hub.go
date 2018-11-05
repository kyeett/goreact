package main

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string](chan bool)

	// Inbound messages from the clients.
	broadcast chan []byte
}

func newHub() *Hub {
	return &Hub{
		broadcast: make(chan []byte),
		clients:   make(map[string](chan bool)),
	}
}

func (h *Hub) run() {
	for {
		select {
		// case client := <-h.register:
		// 	h.clients[client] = true
		// case client := <-h.unregister:
		// 	if _, ok := h.clients[client]; ok {
		// 		delete(h.clients, client)
		// 		close(client.send)
		// 	}
		case <-h.broadcast:
			for k, ch := range h.clients {
				select {
				case ch <- true:
				default:
					close(ch)
					delete(h.clients, k)
				}
			}
		}
	}
}
