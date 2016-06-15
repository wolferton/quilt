/*
Package querymanager provides and supports the QueryManager facility. The QueryManager provides a mechanism for
loading query templates from plain text files and allowing code to combine those templates with parameters to create a
query ready for execution.

The QueryManager is generic and is suitable for managing query templates for any data source that is interacted with via
text queries.
*/
package querymanager

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
)

type QueryManager struct {
	TemplateLocation   string
	VarStart           string
	VarEnd             string
	FrameworkLogger    logger.Logger
	QueryIdPrefix      string
	TrimIdWhiteSpace   bool
	WrapStrings        bool
	StringWrapWith     string
	tokenisedTemplates map[string][]string
}

func (qm *QueryManager) StartComponent() {
	fl := qm.FrameworkLogger
	fl.LogDebug("Starting QueryManager")
	fl.LogDebug(qm.TemplateLocation)

	queryFiles, err := config.FileListFromPath(qm.TemplateLocation)

	for _, file := range queryFiles {
		fl.LogDebug(file)
	}

	if err == nil {

		qm.parseQueryFiles(queryFiles)
		fl.LogDebug("Started QueryManager")
	} else {
		fl.LogFatal("Unable to start QueryManager due to problem loading query files: " + err.Error())
	}

}

func (qm *QueryManager) parseQueryFiles(files []string) {

	fl := qm.FrameworkLogger

	for _, file := range files {

		fl.LogTrace("Parsing " + file)

	}

}
