package querymanager

import (
	"github.com/wolferton/quilt/facility/logger"
	"os"
	"io/ioutil"
	"errors"
)

type QueryManager struct {
	TemplateLocation string
	FrameworkLogger logger.Logger
}

func (qm *QueryManager) StartComponent() {
	fl := qm.FrameworkLogger
	fl.LogDebug("Starting QueryManager")

	queryFiles, err := qm.FileListFromPath(qm.TemplateLocation)

	if err == nil {
		qm.parseQueryFiles(queryFiles)
		fl.LogDebug("Started QueryManager")
	} else {
		fl.LogFatal("Unable to start QueryManager due to problem loading query files: " + err.Error())
	}

}

func (qm *QueryManager) parseQueryFiles(files []string){

}

func (qm *QueryManager) FileListFromPath(path string) ([]string, error) {

	files := make([]string, 0)

	file, err := os.Open(path)

	if err != nil {
		err := errors.New("Unable to open file/dir " + qm.TemplateLocation)
		return files, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {
		err := errors.New("Unable to obtain file info for file/dir " + qm.TemplateLocation)
		return files, err
	}

	if fileInfo.IsDir() {
		contents, err := ioutil.ReadDir(path)

		if err != nil {
			err := errors.New("Unable to read contents of directory " + path)
			return files, err
		}

		files := make([]string, 0)

		for _, info := range contents{

			fileName := info.Name()

			if info.Mode().IsDir() {
				files = append(files, path + "/" + fileName)
			}
		}

	} else {
		files = append(files, file.Name())
	}

	return files, nil
}