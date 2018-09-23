package gossip

import (
	"github.com/golang/glog"
	"github.com/hashicorp/memberlist"
)

type eventDelegate struct{}

func (ed eventDelegate) NotifyJoin(node *memberlist.Node) {
	glog.Infof("NotifyJoin %v", node)
}

func (ed eventDelegate) NotifyLeave(node *memberlist.Node) {
	glog.Infof("NotifyLeave %v", node)
}

func (ed eventDelegate) NotifyUpdate(node *memberlist.Node) {
	glog.Infof("NotifyUpdate %v", node)
}
