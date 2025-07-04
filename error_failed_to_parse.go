package jsonrpc

import "encoding/json"

type ToJSONRPCBytes interface {
	JSONRPCBytes() []byte
}

type ParseError struct {
	JsonRPC  string   `json:"jsonrpc"`
	RpcError RPCError `json:"error"`
	ID       *string  `json:"id"`
}

func NewParseError(details ...Detail) ParseError {
	detailsMap := map[string]interface{}{}
	for _, d := range details {
		detailsMap[d.Key()] = d.Value()
	}
	return ParseError{
		JsonRPC:  "2.0",
		RpcError: RPCError{Code: -32700, Message: "Parse error", Data: detailsMap},
		ID:       nil,
	}
}

func (p ParseError) JSONRPCBytes() []byte {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return b
}

type RPCError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

type Detail interface {
	Key() string
	Value() interface{}
}

type detail struct {
	key   string
	value interface{}
}

func (d *detail) Key() string {
	return d.key
}

func (d *detail) Value() interface{} {
	return d.value
}

func NewDetail(key string, val interface{}) Detail {
	return &detail{
		key:   key,
		value: val,
	}
}
