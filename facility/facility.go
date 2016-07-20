package facility

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/ioc"
	"github.com/wolferton/quilt/logging"
)

type FacilityBuilder interface {
	BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer)
	FacilityName() string
	DependsOnFacilities() []string
}
