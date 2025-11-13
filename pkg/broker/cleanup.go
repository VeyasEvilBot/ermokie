package broker

import "time"

type CleanupPolicy struct {
	ConsumedMessageAge      string        // -1 day
	UnconsumedMessageAge    string        // -1 day
	LogRetentiionTime       string        // -1 day
	InactiveSubscriptionAge string        // -1 day
	CleanupInterval         time.Duration // -1 * time.hour
}
