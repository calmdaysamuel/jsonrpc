package jsonrpc

import "context"

const (
	paramsKey = "jsonrpcContextParams"
)

type Param interface {
	Key() string
	Value() interface{}
}

type safeParam struct {
	key string
	val interface{}
}

type logOnlyParam struct {
	key string
	val interface{}
}

func (u *logOnlyParam) Key() string {
	return u.key

}

func (u *logOnlyParam) Value() interface{} {
	return u.val
}

func (s *safeParam) Key() string {
	return s.key
}

func (s *safeParam) Value() interface{} {
	return s.val
}

func SafeParam(key string, val interface{}) Param {
	return &safeParam{key: key, val: val}
}

func LogOnlyParam(key string, val interface{}) Param {
	return &logOnlyParam{key: key, val: val}
}

func ContextWithParams(ctx context.Context, params ...Param) context.Context {
	val, ok := ctx.Value(paramsKey).(map[string]Param)
	if !ok {
		val = map[string]Param{}
	}

	for _, param := range params {
		val[param.Key()] = param
	}
	return context.WithValue(ctx, paramsKey, val)
}

func ParamsFromContext(ctx context.Context) []Param {
	val, ok := ctx.Value(paramsKey).(map[string]Param)
	if !ok {
		val = map[string]Param{}
	}
	params := make([]Param, 0)
	for _, param := range val {
		params = append(params, param)
	}
	return params
}
