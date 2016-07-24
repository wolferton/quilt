package serviceerror

import (
	"fmt"
	"github.com/wolferton/quilt/logging"
	"github.com/wolferton/quilt/ws"
)

type FrameworkErrorEvent string

const (
	UnableToParseRequest = "UnableToParseRequest"
)

type FrameworkErrorGenerator struct {
	Messages        map[string][]string
	FrameworkLogger logging.Logger
}

func (feg *FrameworkErrorGenerator) Error(e FrameworkErrorEvent, c ws.ServiceErrorCategory, a ...string) *ws.CategorisedError {

	return nil

}

func (feg *FrameworkErrorGenerator) StartComponent() error {

	for k, v := range feg.Messages {
		fmt.Printf("%s %v%d\n", k, v, len(v))

		for _, x := range v {
			fmt.Printf(">>%s<<\n", x)
		}

	}

	return nil

}
