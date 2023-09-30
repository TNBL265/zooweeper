package zooweeper

// SessionTracker tracks session in grouped by tick interval. It always rounds up the tick
// interval to provide a sort of grace period. Sessions are thus expired in batches made up of sessions that expire
// in a given interval.=
type SessionTracker struct {
}
