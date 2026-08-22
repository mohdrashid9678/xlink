package models

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code          string         `json:"code"`
	Message       string         `json:"message"`
	Type          string         `json:"type,omitempty"`
	Title         string         `json:"title,omitempty"`
	Status        int            `json:"status,omitempty"`
	Detail        string         `json:"detail,omitempty"`
	Instance      string         `json:"instance,omitempty"`
	RequestID     string         `json:"request_id,omitempty"`
	InvalidParams []InvalidParam `json:"invalid_params,omitempty"`
}

type InvalidParam struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type ProblemDetails struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Status        int            `json:"status"`
	Detail        string         `json:"detail"`
	Instance      string         `json:"instance,omitempty"`
	Code          string         `json:"code"`
	RequestID     string         `json:"request_id,omitempty"`
	InvalidParams []InvalidParam `json:"invalid_params,omitempty"`
}
