package ioc

type Startable interface {
	StartComponent() error
}

type Stoppable interface {
	StopComponent()
}
