package jsonrpc

type Option = func(opts *serverOpts)

type serverOpts struct {
	maxRequestSize          int64
	batchRequestParallelism int
	maxBatchSize            int
}

func defaultOpts() *serverOpts {
	return &serverOpts{
		maxRequestSize:          1024 * 1024 * 1024, // 1mb
		batchRequestParallelism: 8,
		maxBatchSize:            25,
	}
}

func WithMaxRequestSize(maxSize int64) Option {
	return func(opts *serverOpts) {
		opts.maxRequestSize = maxSize
	}
}

func WithBatchRequestParallelism(parallelism int) Option {
	return func(opts *serverOpts) {
		opts.batchRequestParallelism = parallelism
	}
}

func WithMaxBatchSize(batchSize int) Option {
	return func(opts *serverOpts) {
		opts.maxBatchSize = batchSize
	}
}
