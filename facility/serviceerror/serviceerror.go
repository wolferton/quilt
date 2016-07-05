package serviceerror

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
	"github.com/wolferton/quilt/ws"
	"strings"
)

const (
	serviceErrorManagerComponentName = ioc.FrameworkPrefix + "ServiceErrorManager"
)

type ServiceErrorManager struct {
	errors          map[string]*ws.CategorisedError
	FrameworkLogger logger.Logger
}

func (sem *ServiceErrorManager) Find(code string) *ws.CategorisedError {
	e := sem.errors[code]

	if e == nil {
		sem.FrameworkLogger.LogWarnf("Could not find error with code %s", code)
	}

	return e

}

func (sem *ServiceErrorManager) LoadErrors(definitions []interface{}) {

	l := sem.FrameworkLogger
	sem.errors = make(map[string]*ws.CategorisedError)

	for i, d := range definitions {

		e := d.([]interface{})

		category, err := ws.CodeToCategory(e[0].(string))

		if err != nil {
			l.LogWarnf("Error index %d: %s", i, err.Error())
			continue
		}

		code := e[1].(string)

		if len(strings.TrimSpace(code)) == 0 {
			l.LogWarnf("Error index %d: No code supplied", i)
			continue

		} else if sem.errors[code] != nil {
			l.LogWarnf("Error index %d: Duplicate code", i)
			continue
		}

		message := e[2].(string)

		if len(strings.TrimSpace(message)) == 0 {
			l.LogWarnf("Error index %d: No message supplied", i)
			continue
		}

		ce := ws.NewCategorisedError(category, code, message)

		sem.errors[code] = ce

	}
}

func InitialiseServiceErrorManager(logManager *logger.ComponentLoggerManager, config *config.ConfigAccessor) []*ioc.ProtoComponent {

	if !config.BoolValue("facilities.serviceErrorManager.enabled") {
		return []*ioc.ProtoComponent{}
	} else {

		manager := new(ServiceErrorManager)
		manager.FrameworkLogger = logManager.CreateLogger(serviceErrorManagerComponentName)
		managerProto := ioc.CreateProtoComponent(manager, serviceErrorManagerComponentName)

		definitions := config.StringVal("facilities.serviceErrorManager.errorDefintions")
		errors := config.Array(definitions)

		if errors == nil {
			manager.FrameworkLogger.LogWarnf("No error definitions found at config path %s", definitions)
		} else {
			manager.LoadErrors(errors)
		}

		return []*ioc.ProtoComponent{managerProto}
	}
}
