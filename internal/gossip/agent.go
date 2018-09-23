package gossip

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/hashicorp/memberlist"
)

type Agent struct {
	broadcasts *memberlist.TransmitLimitedQueue
	memberlist *memberlist.Memberlist

	config *Config
}

type Config struct {
	BindAddress net.IP
	BindPort    int

	EncryptKey string

	StartJoin []string

	Server bool
}

func NewAgent(config *Config) (*Agent, error) {
	agent := &Agent{
		config: config,
	}

	hostname, _ := os.Hostname()

	c := memberlist.DefaultWANConfig()
	c.Delegate = &delegate{
		agent: agent,
	}
	c.Events = &eventDelegate{}
	c.BindPort = config.BindPort

	c.Name = fmt.Sprintf("%v-%v", hostname, config.BindPort)

	if len(config.EncryptKey) != 0 {
		c.SecretKey = []byte(config.EncryptKey)
	}

	m, err := memberlist.Create(c)
	if err != nil {
		return nil, err
	}

	agent.memberlist = m

	if len(config.StartJoin) > 0 {
		_, err := m.Join(config.StartJoin)
		if err != nil {
			return nil, err
		}
	}

	agent.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return agent.memberlist.NumMembers()
		},
		RetransmitMult: 3,
	}

	return agent, nil
}

func (a *Agent) StartServer() {
	if !a.config.Server {
		return
	}
}

func (a *Agent) Start() {
	if a.config.Server {
		return
	}

	go func() {
		for {
			time.Sleep(5 * time.Second)
			alive := a.memberlist.NumMembers()
			glog.Infof("Alive members: %v", alive)
		}
	}()
}

func (a *Agent) Stop() error {
	return a.memberlist.Leave(1 * time.Second)
}

type IngressConfiguration struct {
	NGINX []byte `json:"nginx,omitempty"`
	LUA   []byte `json:"lua,omitempty"`
}

func (a *Agent) Broadcast(config *IngressConfiguration) error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}

	a.broadcasts.QueueBroadcast(&broadcast{
		msg: b,
	})

	return nil
}
