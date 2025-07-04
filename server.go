package jsonrpc

import (
	"context"
	"encoding/json"
	"golang.org/x/sync/errgroup"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
)

// New returns a json-rpc server with rational defaults.
// Use the [Register] method to add support for a new rpc call.
// Use the [Start] method to start the server
func New(options ...Option) Server {
	mux := http.NewServeMux()
	opts := defaultOpts()
	for _, option := range options {
		option(opts)
	}
	handler := &jsonRPCServer{
		mux:     mux,
		opts:    opts,
		methods: make(map[string]RPCHandler),
	}
	mux.Handle("/rpc", handler)
	mux.Handle("/rpc/", handler)
	mux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readiness", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/liveliness", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	return handler
}

type RPCHandler interface {
	// MethodName return the name of the RPC server
	MethodName() string
	// Execute computes the result of the rpc call using the provided parameters
	Execute(ctx context.Context, headers http.Header, id *string, params interface{}) (interface{}, error)
	// ParametersValid returns true if the provided parameters can be used with this method
	// The details returned can be used to explain why the parameters are not valid.
	ParametersValid(ctx context.Context, params interface{}) ([]Detail, bool)
}
type Server interface {
	Register(handler RPCHandler)
	Start(port int) error
}

type jsonRPCServer struct {
	opts    *serverOpts
	mux     *http.ServeMux
	methods map[string]RPCHandler
}

func (j *jsonRPCServer) Start(port int) error {
	return http.ListenAndServe(":"+strconv.Itoa(port), j.mux)
}

func (j *jsonRPCServer) Register(handler RPCHandler) {
	if _, ok := j.methods[handler.MethodName()]; ok {
		panic("method all registered")
	}
	j.methods[handler.MethodName()] = handler
}

func (j *jsonRPCServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		err := request.Body.Close()
		if err != nil {
			slog.Error("Failed to close request body")
		}
	}()
	ctx := ContextWithParams(request.Context(), LogOnlyParam("method", request.Method))
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = writer.Write(NewMethodNotFoundError(NewDetail("rationale", "All RPC request should be made with a POST method.")).JSONRPCBytes())
		return
	}

	requestBytes, err := io.ReadAll(io.LimitReader(request.Body, j.opts.maxRequestSize))
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		if _, err := writer.Write(NewParseError(NewDetail("rationale", "Failed to read request body")).JSONRPCBytes()); err != nil {
			slog.Error("Failed to write response body")
		}
		return
	}
	var maybeBatchRequest BatchRequest
	if err := json.Unmarshal(requestBytes, &maybeBatchRequest); err != nil {
		var maybeSingleRequest Request
		if err := json.Unmarshal(requestBytes, &maybeSingleRequest); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			if _, err := writer.Write(NewParseError(NewDetail("rationale", "Failed to parse valid json from request body")).JSONRPCBytes()); err != nil {
				slog.Error("Failed to write response body")
			}
			return
		}
		j.handleSingleRequest(ctx, writer, request, maybeSingleRequest)
		return
	}
	j.handleBatchRequest(ctx, writer, request, maybeBatchRequest)
}

func (j *jsonRPCServer) handleSingleRequest(ctx context.Context, writer http.ResponseWriter, request *http.Request, jsonRequest Request) {
	response, err := j.routeRequest(ctx, request, jsonRequest)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		if gerr, ok := err.(ToJSONRPCBytes); ok {
			if _, err := writer.Write(gerr.JSONRPCBytes()); err != nil {
				slog.Error("Failed to write failed request response body")
			}
			return
		}
		return
	}
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(response.JSONRPCBytes()); err != nil {
		slog.Error("Failed to write failed request response body on successful request")
	}
	return
}

func (j *jsonRPCServer) handleBatchRequest(ctx context.Context, writer http.ResponseWriter, request *http.Request, batchJsonRequest BatchRequest) {
	if len(batchJsonRequest) > j.opts.maxBatchSize {
		writer.WriteHeader(http.StatusBadRequest)
		if _, err := writer.Write(NewInvalidRequestError(nil, NewDetail("rationale", "Too many requests"), NewDetail("maxBatchSize", j.opts.maxBatchSize)).JSONRPCBytes()); err != nil {
			slog.Error("Failed to write response body")
		}
		return
	}
	eg := errgroup.Group{}
	eg.SetLimit(j.opts.batchRequestParallelism)
	lock := sync.Mutex{}
	var responses []interface{}
	for _, r := range batchJsonRequest {
		eg.Go(func() error {
			resp, err := j.routeRequest(ctx, request, r)
			lock.Lock()
			defer lock.Unlock()
			if r.ID != nil {
				if err == nil {
					responses = append(responses, resp)
				} else {
					responses = append(responses, err)
				}
			}
			return nil
		})
	}
	_ = eg.Wait()
	writer.WriteHeader(http.StatusOK)

	b, err := json.Marshal(responses)
	if err == nil {
		_, _ = writer.Write(b)
	}
	return
}

func (j *jsonRPCServer) routeRequest(ctx context.Context, request *http.Request, rpcRequest Request) (_ Response, err error) {
	if rpcRequest.JSONRPC != "2.0" {
		return Response{}, NewInvalidRequestError(rpcRequest.ID, NewDetail("rationale", "Only JSONRPC version 2 is supported"))
	}
	handler, ok := j.methods[rpcRequest.Method]
	if !ok {
		return Response{}, NewMethodNotFoundError()
	}
	slog.Info("Received request", "log.type", "request.v1", "method", rpcRequest.Method)
	if details, ok := handler.ParametersValid(ctx, rpcRequest.Params); !ok {
		return Response{}, NewInvalidRequestError(rpcRequest.ID, details...)
	}
	result, err := handler.Execute(ctx, request.Header, rpcRequest.ID, rpcRequest.Params)
	if err != nil {
		if _, ok := err.(ToJSONRPCBytes); ok {
			return Response{}, err
		}
		return Response{}, FromStandardError(rpcRequest.ID, err)
	}
	return NewResponse(rpcRequest.ID, result), nil
}
