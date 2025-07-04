package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

type adder struct {
}

func (a *adder) MethodName() string {
	return "add"
}

func (a *adder) Execute(ctx context.Context, headers http.Header, id *string, params interface{}) (interface{}, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return nil, errors.New("parameters MUST be an int64 array")
	}
	var intArray []int64
	if err := json.Unmarshal(b, &intArray); err != nil {
		return nil, errors.New("parameters MUST be an int64 array")
	}
	sum := int64(0)
	for _, i := range intArray {
		sum += i
	}
	return sum, nil
}

func (a *adder) ParametersValid(ctx context.Context, params interface{}) ([]Detail, bool) {
	b, err := json.Marshal(params)
	if err != nil {
		return []Detail{NewDetail("rationale", "parameters MUST be an int64 array")}, false
	}
	var intArray []int64
	if err := json.Unmarshal(b, &intArray); err != nil {
		return []Detail{NewDetail("rationale", "parameters MUST be an int64 array")}, false
	}
	return nil, true
}

func TestNewServer(t *testing.T) {
	server := New()
	server.Register(&adder{})
	go require.NoErrorf(t, server.Start(3124), "server failed to start up successfully")
}
