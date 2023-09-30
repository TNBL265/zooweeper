package zooweeper

// FollowerRequestProcessor forwards any requests that modify the state of the system to the Leader.
type FollowerRequestProcessor struct {
}

func (f FollowerRequestProcessor) ProcessRequest(request *Request) error {
	//TODO implement me
	panic("implement me")
}

// SendAckRequestProcessor implements request processing logic for followers.
type SendAckRequestProcessor struct {
}

func (s SendAckRequestProcessor) ProcessRequest(request *Request) error {
	//TODO implement me
	panic("implement me")
}
