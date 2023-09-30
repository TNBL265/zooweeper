package zooweeper

// RequestProcessor are chained together to process transactions. Requests are  always processed in order.
// The standalone server, follower, and leader all have slightly different RequestProcessors changed together.
//
// Requests always move forward through the chain of RequestProcessors. Requests are passed to a RequestProcessor
// through processRequest(). Generally method will always be invoked by a single thread.
//
// When shutdown is called, the request RequestProcessor should also shutdown any RequestProcessors that it is connected to.
type RequestProcessor interface {
	// ProcessRequest processes a request and returns an error if any.
	ProcessRequest(request *Request) error
}

// Request is the structure that represents a request moving through a chain of RequestProcessor.
// There are various pieces of information that is tacked onto the request as it is processed.
type Request struct {
}
