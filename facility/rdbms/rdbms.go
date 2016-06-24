package rdbms

import (
	"database/sql"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/facility/querymanager"
)

type DatabaseProvider interface {
	Database() (*sql.DB, error)
}

type DatabaseAccessor struct {
	Provider                      DatabaseProvider
	QueryManager                  *querymanager.QueryManager
	FrameworkLogger               logger.Logger
	DatabaseProviderComponentName string
}

func (da *DatabaseAccessor) InsertQueryIdParamObject(queryId string, params interface{}) error {

	db, err := da.Provider.Database()

	err = db.Ping()

	return err
}
