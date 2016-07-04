package serviceerror

import (
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ws"
)

type ServiceErrorManager struct {
	errors          map[string]*ws.CategorisedError
	FrameworkLogger logger.Logger
	ConfigLocation  string
}
