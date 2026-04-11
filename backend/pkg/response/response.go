package response

type Response[T any] struct {
	Data      T           `json:"data,omitempty"`
	Error     *ErrorField `json:"error,omitempty"`
	Page      *PageInfo   `json:"page,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type ErrorField struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type PageInfo struct {
	PageNum  int   `json:"page_num"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func Success[T any](data T) Response[T] {
	return Response[T]{Data: data}
}

func SuccessWithPage[T any](data T, pageNum, pageSize int, total int64) Response[T] {
	return Response[T]{
		Data: data,
		Page: &PageInfo{
			PageNum:  pageNum,
			PageSize: pageSize,
			Total:    total,
		},
	}
}

func Fail(code, message string) Response[any] {
	return Response[any]{
		Error: &ErrorField{
			Code:    code,
			Message: message,
		},
	}
}

func FailWithDetails(code, message string, details any) Response[any] {
	return Response[any]{
		Error: &ErrorField{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
