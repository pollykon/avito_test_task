package log

import (
	"time"
)

type GetCSVRequest struct {
	UserID    int64
	From      time.Time
	To        time.Time
	Separator string
}
