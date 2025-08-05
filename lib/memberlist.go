package lib

import (
	"log/slog"
	"os"
	"time"

	"github.com/hashicorp/memberlist"
)

func InitMemberList(knownMembers []string, port int, proxyPort string, manager *QueueManager) *memberlist.Memberlist {
	config := memberlist.DefaultLANConfig()
	config.BindPort = port
	config.AdvertisePort = port
	config.Delegate = NirnDelegate{
		proxyPort: proxyPort,
	}

	config.Events = manager.GetEventDelegate()

	// DEBUG CODE
	if os.Getenv("NODE_NAME") != "" {
		config.Name = os.Getenv("NODE_NAME")
		config.DeadNodeReclaimTime = 1 * time.Nanosecond
	}

	list, err := memberlist.Create(config)
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	manager.SetCluster(list, proxyPort)

	_, err = list.Join(knownMembers)
	if err != nil {
		slog.Error("Failed to join existing cluster, ok if this is the first node", "error", err)
	}

	var members string
	for _, member := range list.Members() {
		members += member.Name + " "
	}

	slog.Info("Connected to cluster nodes", "members", members)
	return list
}
