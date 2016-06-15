package querymanager

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
)

type QueryManager struct {
	TemplateLocation string
	FrameworkLogger  logger.Logger
}

func (qm *QueryManager) StartComponent() {
	fl := qm.FrameworkLogger
	fl.LogDebug("Starting QueryManager")

	queryFiles, err := config.FileListFromPath(qm.TemplateLocation)

	if err == nil {
		qm.parseQueryFiles(queryFiles)
		fl.LogDebug("Started QueryManager")
	} else {
		fl.LogFatal("Unable to start QueryManager due to problem loading query files: " + err.Error())
	}

}

func (qm *QueryManager) parseQueryFiles(files []string) {

}
