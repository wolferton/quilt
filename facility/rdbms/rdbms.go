package rdbms

import (
	"database/sql"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/facility/querymanager"
)

type DatabaseProvider interface {
	Database() (*sql.DB, error)
}

type RdbmsClientManager interface {
	Client() *RdbmsClient
	ClientFromContext(context interface{}) *RdbmsClient
}

type DefaultRdbmsClientManager struct {
	Provider                      DatabaseProvider
	DatabaseProviderComponentName string
	QueryManager                  *querymanager.QueryManager
	db                            *sql.DB
	FrameworkLogger               logger.Logger
}

func (drcm *DefaultRdbmsClientManager) Client() *RdbmsClient {
	return newRdbmsClient(drcm.db, drcm.QueryManager)
}

func (drcm *DefaultRdbmsClientManager) ClientFromContext(context interface{}) *RdbmsClient {
	return drcm.Client()
}

func (drcm *DefaultRdbmsClientManager) StartComponent() error {

	db, err := drcm.Provider.Database()

	if err != nil {
		return err

	} else {
		drcm.db = db
		return nil
	}

}

func newRdbmsClient(database *sql.DB, querymanager *querymanager.QueryManager) *RdbmsClient {
	rc := new(RdbmsClient)
	rc.db = database
	rc.queryManager = querymanager

	return rc
}

type RdbmsClient struct {
	db           *sql.DB
	queryManager *querymanager.QueryManager
}

func (rc *RdbmsClient) InsertQueryIdParamMap(queryId string, params map[string]interface{}) (sql.Result, error) {

	query, err := rc.queryManager.SubstituteMap(queryId, params)

	if err != nil {
		return nil, err
	}

	result, err := rc.db.Exec(query)

	return result, err
}

func (rc *RdbmsClient) InsertQueryIdParamMapReturnedId(queryId string, params map[string]interface{}) (int, error) {

	query, err := rc.queryManager.SubstituteMap(queryId, params)

	if err != nil {
		return 0, err
	}

	var id int

	err = rc.db.QueryRow(query).Scan(&id)

	return id, err
}

func (rc *RdbmsClient) SelectQueryIdParamMap(queryId string, params map[string]interface{}) (*sql.Rows, error) {
	query, err := rc.queryManager.SubstituteMap(queryId, params)

	if err != nil {
		return nil, err
	}

	return rc.db.Query(query)

}
