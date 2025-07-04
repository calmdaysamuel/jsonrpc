package jsonrpc

type BatchRequest = []Request

type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      *string     `json:"id,omitempty"`
	Params  interface{} `json:"params,omitempty"`
}
