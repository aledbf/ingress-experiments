package server

type fromAgentFn func(message []byte)

type Hub struct {
	agents map[*Agent]bool

	broadcast chan []byte

	register   chan *Agent
	unregister chan *Agent

	fromAgentFn fromAgentFn
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Agent),
		unregister: make(chan *Agent),
		agents:     make(map[*Agent]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case agent := <-h.register:
			h.agents[agent] = true
		case agent := <-h.unregister:
			if _, ok := h.agents[agent]; ok {
				delete(h.agents, agent)
				close(agent.send)
			}
		case message := <-h.broadcast:
			for agent := range h.agents {
				select {
				case agent.send <- message:
				default:
					close(agent.send)
					delete(h.agents, agent)
				}
			}
		}
	}
}

func (h *Hub) SetFromAgentFn(fn fromAgentFn) {
	h.fromAgentFn = fn
}

func (h *Hub) SendUpdate(data []byte) {
	h.broadcast <- data
}

func (h *Hub) MessageFromAgent(message []byte) {
	h.fromAgentFn(message)
}
