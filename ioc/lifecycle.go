package ioc

type Startable interface {
    StartComponent()
}

type Stoppable interface {
    StopComponent()
}