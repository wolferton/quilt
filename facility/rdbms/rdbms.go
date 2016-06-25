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

func (da *DatabaseAccessor) InsertQueryIdParamMap(queryId string, params map[string]interface{}) (sql.Result, error) {

	query, err := da.QueryManager.SubstituteMap(queryId, params)

	if err != nil {
		return nil, err
	}

	da.FrameworkLogger.LogInfo(query)

	db, err := da.Provider.Database()

	result, err := db.Exec(query)

	return result, err
}
