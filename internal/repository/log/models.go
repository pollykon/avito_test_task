package log

import "time"

type Log struct {
	ID         int64
	UserID     int64
	SegmentID  string
	Operation  string
	InsertTime time.Time
}
