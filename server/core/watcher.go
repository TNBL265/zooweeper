package zooweeper

// Watcher specifies the public interface an event handler class must implement. A ZooWeeperClient will get
// various events from the ZooWeeperServer it connects to. An application using such a client handles these events by
// registering a callback object with the client. The callback object is expected to be an instance of a class that
// implements Watcher interface.
type Watcher interface {
}
