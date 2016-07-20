package querymanager

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
)

const QueryManagerComponentName = ioc.FrameworkPrefix + "QueryManager"
const QueryManagerFacilityName = "QueryManager"

type QueryManagerFacilityBuilder struct {
}

func (qmfb *QueryManagerFacilityBuilder) BuildAndRegister(lm *logger.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) {

	queryManager := new(QueryManager)
	ca.Populate("QueryManager", queryManager)

	cn.WrapAndAddProto(QueryManagerComponentName, queryManager)

}

func (qmfb *QueryManagerFacilityBuilder) FacilityName() string {
	return QueryManagerFacilityName
}

func (qmfb *QueryManagerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
