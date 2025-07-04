package jsonrpc

import "encoding/json"

type MethodNotFoundError struct {
	JsonRPC  string   `json:"jsonrpc"`
	RpcError RPCError `json:"error"`
	ID       *string  `json:"id"`
}

func (p MethodNotFoundError) Error() string {
	return p.RpcError.Message
}

func NewMethodNotFoundError(details ...Detail) MethodNotFoundError {
	detailsMap := map[string]interface{}{}
	for _, d := range details {
		detailsMap[d.Key()] = d.Value()
	}
	return MethodNotFoundError{
		JsonRPC:  "2.0",
		RpcError: RPCError{Code: -32601, Message: "Method not found"},
		ID:       nil,
	}
}

func (p MethodNotFoundError) JSONRPCBytes() []byte {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return b
}
