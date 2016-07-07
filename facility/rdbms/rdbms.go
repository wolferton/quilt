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
