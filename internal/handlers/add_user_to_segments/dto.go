package add_user_to_segments

type HandlerRequest struct {
	UserID       int64    `json:"userId"`
	SegmentSlugs []string `json:"slugs"`
	TTLHours     *int64   `json:"ttl"`
}

type HandlerResponse struct {
	Status int                   `json:"status"`
	Error  *HandlerResponseError `json:"error,omitempty"`
}

type HandlerResponseError struct {
	Message string `json:"message"`
}
