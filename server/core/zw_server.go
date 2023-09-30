package zooweeper

// ZWServer implements a simple standalone ZooKeeperServer. Request Processors flow:
// PrepRequestProcessor -> SyncRequestProcessor -> FinalRequestProcessor
type ZWServer struct {
}
