package main

import (
	"flag"
	"fmt"
	"github.com/ravilr/redis-failover/failover"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var configFile = flag.String("config", "", "failover config file")
var addr = flag.String("addr", "", "failover http listen addr")
var checkInterval = flag.Int("check_interval", 1000, "check master alive every n millisecond")
var maxDownTime = flag.Int("max_down_time", 5, "max down time for a master, after that, we will do failover")

var masters = flag.String("masters", "", "redis master need to be monitored, seperated by comma")
var mastersState = flag.String("masters_state", "", "new or existing for raft, if new, we will depracted old saved masters")

var broker = flag.String("broker", "raft", "broker for cluster, default is raft")

var raftDataDir = flag.String("raft_data_dir", "", "raft data store path")
var raftLogDir = flag.String("raft_log_dir", "", "raft log store path")
var raftAddr = flag.String("raft_addr", "", "raft listen addr, if empty, we will disable raft")
var raftCluster = flag.String("raft_cluster", "", "raft cluster, seperated by comma")
var raftClusterState = flag.String("raft_cluster_state", "", "new or existing, if new, we will deprecate old saved cluster and use new")

func main() {
	flag.Parse()

	var c *failover.Config
	var err error
	if len(*configFile) > 0 {
		c, err = failover.NewConfigWithFile(*configFile)
		if err != nil {
			fmt.Printf("load failover config %s err %v", *configFile, err)
			return
		}
	} else {
		fmt.Printf("no config file, use default config")
		c = new(failover.Config)
		c.Addr = ":11000"
		c.CheckInterval = 1000
		c.MaxDownTime = 3
	}

	if len(*addr) > 0 {
		c.Addr = *addr
	}

	if *checkInterval > 0 {
		c.CheckInterval = *checkInterval
	}

	if *maxDownTime > 0 {
		c.MaxDownTime = *maxDownTime
	}

	if len(*raftAddr) > 0 {
		c.Raft.Addr = *raftAddr
	}

	if len(*raftDataDir) > 0 {
		c.Raft.DataDir = *raftDataDir
	}

	if len(*raftLogDir) > 0 {
		c.Raft.LogDir = *raftLogDir
	}

	seps := strings.Split(*raftCluster, ",")
	if len(seps) > 0 && len(seps[0]) > 0 {
		c.Raft.Cluster = seps
	}

	if len(*raftClusterState) > 0 {
		c.Raft.ClusterState = *raftClusterState
	}

	if len(*broker) > 0 {
		c.Broker = *broker
	}

	seps = strings.Split(*masters, ",")
	if len(seps) > 0 && len(seps[0]) > 0 {
		c.Masters = seps
	}

	if len(*mastersState) > 0 {
		c.MastersState = *mastersState
	}

	app, err := failover.NewApp(c)
	if err != nil {
		fmt.Printf("new failover app error %v", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-sc
		app.Close()
	}()

	app.Run()
}
