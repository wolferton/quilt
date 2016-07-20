package facility

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
)

type FacilityBuilder interface {
	BuildAndRegister(lm *logger.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer)
	FacilityName() string
	DependsOnFacilities() []string
}
