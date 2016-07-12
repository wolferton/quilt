package ws

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func NewWsQueryParams(values url.Values) *WsQueryParams {

	qp := new(WsQueryParams)
	qp.values = values

	return qp

}

type WsQueryParams struct {
	values url.Values
}

func (qp *WsQueryParams) Exists(key string) bool {
	return qp.values[key] != nil
}

func (qp *WsQueryParams) MultipleValues(key string) bool {

	value := qp.values[key]

	return value != nil && len(value) > 1

}

func (qp *WsQueryParams) StringValue(key string) (string, error) {

	s := qp.values[key]

	if s == nil {
		return "", qp.noVal(key)
	}

	return s[len(s)-1], nil

}

func (qp *WsQueryParams) BoolValue(key string) (string, error) {

	v := qp.values[key]

	if v == nil {
		return false, qp.noVal(key)
	}

	b, err := strconv.ParseBool(v)

	return b, err

}

func (qp *WsQueryParams) noVal(key string) error {
	message := fmt.Sprintf("No value available for key %s", key)
	return errors.New(message)
}
