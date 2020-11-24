package cherry

import (
	"flag"
	"github.com/cherry-game/cherry/cluster"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"time"
)

func GetDefaultValue() (configPath, profileName, nodeId string) {
	flag.StringVar(&configPath, "path", "./config", "-path=~/git/project/config")
	flag.StringVar(&profileName, "profile", "local", "-profile=local")
	flag.StringVar(&nodeId, "node", "web-1", "-node=web-1")
	flag.Parse()
	return configPath, profileName, nodeId
}

// CreateApp create new application instance
func CreateApp(configPath, profileName, nodeId string) *Application {
	err := cherryProfile.Init(configPath, profileName)
	if err != nil {
		panic(err)
	}

	//set logger
	cherryLogger.SetLogger(cherryProfile.Config())

	//print version info
	cherryConst.PrintVersion()

	//load nodes from config file
	cherryCluster.LoadNodes(cherryProfile.Config())

	nodeType, err := cherryCluster.Nodes().GetType(nodeId)
	if err != nil {
		cherryLogger.Panic(err)
		return nil
	}

	app := &Application{
		nodeId:    nodeId,
		nodeType:  nodeType,
		startTime: time.Now().Unix(),
		running:   false,
		die:       make(chan bool),
	}
	return app
}
