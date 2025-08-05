package lib

import (
	"log/slog"

	"github.com/hashicorp/memberlist"
)

type NirnEvents struct {
	memberlist.EventDelegate
	OnJoin  func(node *memberlist.Node)
	OnLeave func(node *memberlist.Node)
}

func formatNodeInfo(node *memberlist.Node) string {
	return node.Name + " - " + node.Address() + " - listenport: " + string(node.Meta)
}

func (d NirnEvents) NotifyJoin(node *memberlist.Node) {
	slog.Info("Node joined the cluster: " + formatNodeInfo(node))
	d.OnJoin(node)
}
func (d NirnEvents) NotifyLeave(node *memberlist.Node) {
	slog.Info("Node left the cluster: " + formatNodeInfo(node))
	d.OnLeave(node)
}
func (d NirnEvents) NotifyUpdate(node *memberlist.Node) {}
