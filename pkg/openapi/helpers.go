package openapi

// CompletePromiseRequestWrapper makes life easier since id is not part of the body.
type CompletePromiseRequestWrapper struct {
	Id      *string
	Request interface{}
}
