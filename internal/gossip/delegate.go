package gossip

import (
	"encoding/json"

	"github.com/golang/glog"
)

type delegate struct {
	agent *Agent
}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *delegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	glog.Infof("NotifyMsg %v", string(b))
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return d.agent.broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	b, _ := json.Marshal(struct{}{})
	return b
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}

	if !join {
		return
	}

	glog.Infof("MergeRemoteState %v", string(buf))
}
