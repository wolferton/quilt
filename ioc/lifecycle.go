package ioc

type Startable interface {
	StartComponent() error
}

type Stoppable interface {
	StopComponent()
}

type Accessible interface {
	MakeAvailable()
}
