package delete_user_from_segment

type HandlerRequest struct {
	UserID       int64    `json:"userId"`
	SegmentSlugs []string `json:"slugs"`
}

type HandlerResponse struct {
	Status int                   `json:"status"`
	Error  *HandlerResponseError `json:"error,omitempty"`
}

type HandlerResponseError struct {
	Message string `json:"message"`
}
