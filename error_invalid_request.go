package jsonrpc

import "encoding/json"

type InvalidRequestError struct {
	JsonRPC  string   `json:"jsonrpc"`
	RpcError RPCError `json:"error"`
	ID       *string  `json:"id"`
}

func (p InvalidRequestError) Error() string {
	return p.RpcError.Message
}

func NewInvalidRequestError(id *string, details ...Detail) InvalidRequestError {
	detailsMap := map[string]interface{}{}
	for _, d := range details {
		detailsMap[d.Key()] = d.Value()
	}
	return InvalidRequestError{
		JsonRPC:  "2.0",
		RpcError: RPCError{Code: -32600, Message: "Invalid Request"},
		ID:       id,
	}
}

func (p InvalidRequestError) JSONRPCBytes() []byte {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return b
}
