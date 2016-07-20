package rdbms

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/facility/querymanager"
	"github.com/wolferton/quilt/ioc"
)

const rdbmsClientManagerName = ioc.FrameworkPrefix + "RdbmsClientManager"

type RdbmsAccessFacilityBuilder struct {
}

func (rafb *RdbmsAccessFacilityBuilder) BuildAndRegister(lm *logger.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) {

	manager := new(DefaultRdbmsClientManager)
	ca.Populate("RdbmsAccess", manager)

	proto := ioc.CreateProtoComponent(manager, rdbmsClientManagerName)

	proto.AddDependency("Provider", manager.DatabaseProviderComponentName)
	proto.AddDependency("QueryManager", querymanager.QueryManagerComponentName)

	cn.AddProto(proto)

}

func (rafb *RdbmsAccessFacilityBuilder) FacilityName() string {
	return "RdbmsAccess"
}

func (rafb *RdbmsAccessFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
