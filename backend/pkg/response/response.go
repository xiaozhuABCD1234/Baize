package response

// Response 标准API响应结构
type Response struct {
	Data      interface{} `json:"data,omitempty" swaggertype:"object"`
	Error     *ErrorField `json:"error,omitempty"`
	Page      *PageInfo   `json:"page,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorField 错误详情结构
type ErrorField struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty" swaggertype:"object"`
}

// PageInfo 分页信息结构
type PageInfo struct {
	PageNum  int   `json:"page_num"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// Success 返回成功响应
func Success(data interface{}) Response {
	return Response{Data: data}
}

// SuccessWithPage 返回带分页的成功响应
func SuccessWithPage(data interface{}, pageNum, pageSize int, total int64) Response {
	return Response{
		Data: data,
		Page: &PageInfo{
			PageNum:  pageNum,
			PageSize: pageSize,
			Total:    total,
		},
	}
}

// Fail 返回失败响应
func Fail(code, message string) Response {
	return Response{
		Error: &ErrorField{
			Code:    code,
			Message: message,
		},
	}
}

// FailWithDetails 返回带详细信息的失败响应
func FailWithDetails(code, message string, details any) Response {
	return Response{
		Error: &ErrorField{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
