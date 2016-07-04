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

	db, err := da.Provider.Database()

	result, err := db.Exec(query)

	return result, err
}

func (da *DatabaseAccessor) InsertQueryIdParamMapReturnedId(queryId string, params map[string]interface{}) (int, error) {

	query, err := da.QueryManager.SubstituteMap(queryId, params)

	if err != nil {
		return 0, err
	}

	db, err := da.Provider.Database()
	var id int

	err = db.QueryRow(query).Scan(&id)

	return id, err
}

func (da *DatabaseAccessor) SelectQueryIdParamMap(queryId string, params map[string]interface{}) (*sql.Rows, error) {
	query, err := da.QueryManager.SubstituteMap(queryId, params)

	if err != nil {
		return nil, err
	}

	db, err := da.Provider.Database()

	return db.Query(query)

}
