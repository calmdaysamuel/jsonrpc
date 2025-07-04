package jsonrpc

import "encoding/json"

type GeneralError struct {
	JsonRPC  string   `json:"jsonrpc"`
	RpcError RPCError `json:"error"`
	ID       *string  `json:"id"`
}

func (g GeneralError) Error() string {
	return g.RpcError.Message
}
func FromStandardError(id *string, err error) GeneralError {
	return GeneralError{
		JsonRPC:  "2.0",
		RpcError: RPCError{Code: 0, Message: err.Error()},
		ID:       id,
	}
}

func NewGeneralError(id *string, message string, code int, details ...Detail) GeneralError {
	detailsMap := map[string]interface{}{}
	for _, d := range details {
		detailsMap[d.Key()] = d.Value()
	}
	return GeneralError{
		JsonRPC:  "2.0",
		RpcError: RPCError{Code: code, Message: message, Data: detailsMap},
		ID:       id,
	}
}

func (p GeneralError) JSONRPCBytes() []byte {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return b
}
