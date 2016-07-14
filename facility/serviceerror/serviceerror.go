package serviceerror

import (
	"fmt"
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
	"github.com/wolferton/quilt/ws"
	"strings"
)

const (
	serviceErrorManagerComponentName   = ioc.FrameworkPrefix + "ServiceErrorManager"
	serviceErrorDecoratorComponentName = ioc.FrameworkPrefix + "ServiceErrorSourceDecorator"
)

type ServiceErrorManager struct {
	errors          map[string]*ws.CategorisedError
	FrameworkLogger logger.Logger
	PanicOnMissing  bool
}

func (sem *ServiceErrorManager) Find(code string) *ws.CategorisedError {
	e := sem.errors[code]

	if e == nil {
		message := fmt.Sprintf("ServiceErrorManager could not find error with code %s", code)

		if sem.PanicOnMissing {
			panic(message)

		} else {
			sem.FrameworkLogger.LogWarnf(message)

		}

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

func InitialiseServiceErrorManager(logManager *logger.ComponentLoggerManager, config *config.ConfigAccessor, container *ioc.ComponentContainer) {

	manager := new(ServiceErrorManager)
	manager.PanicOnMissing = config.BoolValue("ServiceErrorManager.PanicOnMissing")
	container.WrapAndAddProto(serviceErrorManagerComponentName, manager)

	decorator := new(ServiceErrorConsumerDecorator)
	decorator.ErrorSource = manager
	container.WrapAndAddProto(serviceErrorDecoratorComponentName, decorator)

	definitions := config.StringVal("ServiceErrorManager.ErrorDefinitions")
	errors := config.Array(definitions)

	if errors == nil {
		manager.FrameworkLogger.LogWarnf("No error definitions found at config path %s", definitions)
	} else {
		manager.LoadErrors(errors)
	}
}

type ServiceErrorConsumerDecorator struct {
	ErrorSource *ServiceErrorManager
}

func (secd *ServiceErrorConsumerDecorator) OfInterest(component *ioc.Component) bool {
	_, found := component.Instance.(ws.ServiceErrorConsumer)

	return found
}

func (secd *ServiceErrorConsumerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	c := component.Instance.(ws.ServiceErrorConsumer)
	c.ProvideErrorFinder(secd.ErrorSource)
}

type FrameworkServiceErrorFinder interface {
	UnmarshallError() *ws.CategorisedError
}
