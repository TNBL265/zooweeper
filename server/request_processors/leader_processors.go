package zooweeper

// ProposalRequestProcessor simply forwards requests to an AckRequestProcessor and SyncRequestProcessor.
type ProposalRequestProcessor struct {
}

func (p ProposalRequestProcessor) ProcessRequest(request *Request) error {
	//TODO implement me
	panic("implement me")
}

// AckRequestProcessor simply forwards a request from a previous stage to the leader as an ACK.
type AckRequestProcessor struct {
}

func (a AckRequestProcessor) ProcessRequest(request *Request) error {
	//TODO implement me
	panic("implement me")
}

// ToBeAppliedRequestProcessor simply maintains the toBeApplied list. For this to work next must be a
// FinalRequestProcessor and FinalRequestProcessor.processRequest MUST process the request synchronously!
type ToBeAppliedRequestProcessor struct {
}

func (t ToBeAppliedRequestProcessor) ProcessRequest(request *Request) error {
	//TODO implement me
	panic("implement me")
}
