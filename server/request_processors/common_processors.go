package zooweeper

// Common Request Processors used in both leader and followers

// PrepRequestProcessor is generally at the start of a RequestProcessor change. It sets up any transactions associated
// with requests that change the state of the system. It counts on ZooKeeperServer to update outstandingRequests,
// so that it can take into account transactions that are in the queue to be applied when generating a transaction.
type PrepRequestProcessor struct {
}

func (p PrepRequestProcessor) ProcessRequest(request *Request) error {
	//TODO implement me
	panic("implement me")
}

// CommitProcessor matches the incoming committed requests with the locally submitted requests. The trick is that
// locally submitted requests that change the state of the system will come back as incoming committed requests,
// so we need to match them up.
type CommitProcessor struct {
}

func (c CommitProcessor) ProcessRequest(request *Request) error {
	//TODO implement me
	panic("implement me")
}

// SyncRequestProcessor logs requests to disk. It batches the requests to do the io efficiently. The request is not
// passed to the next RequestProcessor until its log has been synced to disk.
type SyncRequestProcessor struct {
}

func (s SyncRequestProcessor) ProcessRequest(request *Request) error {
	//TODO implement me
	panic("implement me")
}

// FinalRequestProcessor applies any transaction associated with a request and services any queries. It is always at
// the end of a RequestProcessor chain (hence the name), so it does not have a nextProcessor member.
//
// This RequestProcessor counts on ZWServer to populate the outstandingRequests member of ZWServer.
type FinalRequestProcessor struct {
}

func (f FinalRequestProcessor) ProcessRequest(request *Request) error {
	//TODO implement me
	panic("implement me")
}
