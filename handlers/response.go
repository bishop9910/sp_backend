package handlers

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error" example:"invalid request"`
	Message string `json:"message" example:"详细错误信息"`
}
