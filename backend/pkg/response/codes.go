package response

const (
	UserNotFound    = "USER_001"
	UserEmailExists = "USER_002"
	InvalidPassword = "USER_003"
	SamePassword    = "USER_004"
	UserNameExists  = "USER_005"

	TokenInvalid     = "AUTH_001"
	TokenExpired     = "AUTH_002"
	TokenTypeInvalid = "AUTH_003"

	BadRequest       = "COMMON_001"
	InternalError    = "COMMON_002"
	ResourceConflict = "COMMON_003"
)
