package zooweeper

// ZWClient is the main class of ZW client library. To use a ZooWeeperClient service, an application must
// first instantiate an object of ZooWeeperClient class. All the iterations will be done by calling the methods of
// ZooWeeperClient class.
//
// To create a ZooWeeperClient object, the application needs to pass a string  containing a list of host:port pairs,
// each corresponding to a ZooWeeperServer ; a sessionTimeout; and an object of Watcher type.
//
// The client object will pick an arbitrary server and try to connect to it. If failed, it will try the next one in
// the list, until a connection is established, or all the servers have been tried.
//
// Once a connection to a server is established, a session ID is assigned to the client. The client will send heartbeats
// to the server periodically to keep the session valid.
//
// The application can call ZW APIs through a client as long as the session ID of the client remains valid.
//
// If for some reason, the client fails to send heart beats to the server for a prolonged period of time (exceeding the
// sessionTimeout value, for instance), the server will expire the session, and the session ID will become invalid.
// The client object will no longer be usable. To make ZooWeeperClient API calls, the  application must create a new
// client object.
//
// If the ZooWeeperServer the client currently connects to fails or otherwise does not respond, the client will
// automatically try to connect to another server before its session ID expires. If successful, the application can
// continue to use the client.
//
// Some successful ZW API calls can leave watches on the "data nodes" in the ZooWeeperServer. Other successful ZW API
// calls can trigger those watches. Once a watch is triggered, an event will be delivered to the client which left the
// watch at the first place. Each watch can be triggered only once. Thus, up to one event will be delivered to a client
// for every watch it leaves.
//
// A client needs an object of a class implementing Watcher interface for processing the events delivered to the client.
//
// When a client drops current connection and re-connects to a server, all the existing watches are considered as being
// triggered but the undelivered events are lost. To emulate this, the client will generate a special event to tell
// the event handler a connection has been dropped.
type ZWClient struct {
}
