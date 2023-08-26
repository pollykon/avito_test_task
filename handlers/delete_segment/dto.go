package delete_segment

type HandlerRequest struct {
	SegmentSlug string `json:"slug"`
}

type HandlerResponse struct {
	Status int                   `json:"status"`
	Error  *HandlerResponseError `json:"error,omitempty"`
}

type HandlerResponseError struct {
	Message string `json:"message"`
}
