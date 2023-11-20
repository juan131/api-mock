package api

const (
	// 1xxx errors
	requestBase           = 1000
	CodeInvalidBody       = requestBase + 1
	CodeNotFound          = requestBase + 2
	CodeMethodNotAllowed  = requestBase + 3
	CodeRateLimitExceeded = requestBase + 4
	CodeFailedRequest     = requestBase + 5
)
