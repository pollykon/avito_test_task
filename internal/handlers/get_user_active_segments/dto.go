package get_user_active_segments

type HandlerRequest struct {
	UserID int64 `json:"userId"`
}

type HandlerResponse struct {
	Status   int                   `json:"status"`
	Error    *HandlerResponseError `json:"error,omitempty"`
	Response []string              `json:"response"`
}

type HandlerResponseError struct {
	Message string `json:"message"`
}
