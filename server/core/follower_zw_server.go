package zooweeper

// FollowerZWServer Request Processors flow: FollowerRequestProcessor -> CommitProcessor -> FinalRequestProcessor
// A SyncRequestProcessor is also spawn off to log proposals from the leader.
type FollowerZWServer struct {
}
