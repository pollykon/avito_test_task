package get_logs

type HandlerRequest struct {
	UserID    int64   `json:"userId"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Separator *string `json:"separator"`
}

type HandlerResponse struct {
	Status int                   `json:"status"`
	Error  *HandlerResponseError `json:"error,omitempty"`
	URI    string                `json:"uri,omitempty"`
}

type HandlerResponseError struct {
	Message string `json:"message"`
}
