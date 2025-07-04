package examples

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/calmdaysamuel/jsonrpc"
	"log"
	"net/http"
)

func main() {
	jsonRPCAdder := Adder()
	if err := jsonRPCAdder.Start(1234); err != nil {
		log.Fatal(err)
	}
}
func Adder() jsonrpc.Server {
	s := jsonrpc.New()
	s.Register(&adder{})
	return s
}

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

func (a *adder) ParametersValid(ctx context.Context, params interface{}) ([]jsonrpc.Detail, bool) {
	b, err := json.Marshal(params)
	if err != nil {
		return []jsonrpc.Detail{jsonrpc.NewDetail("rationale", "parameters MUST be an int64 array")}, false
	}
	var intArray []int64
	if err := json.Unmarshal(b, &intArray); err != nil {
		return []jsonrpc.Detail{jsonrpc.NewDetail("rationale", "parameters MUST be an int64 array")}, false
	}
	return nil, true
}
