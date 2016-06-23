package rdbms

import (
	"database/sql"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/facility/querymanager"
)

type DatabaseProvider interface {
	Database() *sql.DB
}

type DatabaseAccessor struct {
	Provider                      DatabaseProvider
	QueryManager                  *querymanager.QueryManager
	FrameworkLogger               logger.Logger
	DatabaseProviderComponentName string
}
