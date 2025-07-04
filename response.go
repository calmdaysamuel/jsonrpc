package jsonrpc

import "encoding/json"

type BatchResponse = []Response

type Response struct {
	JsonRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      *string     `json:"id"`
}

func NewResponse(id *string, result interface{}) Response {
	return Response{
		JsonRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

func (p Response) JSONRPCBytes() []byte {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return b
}
