package lib

func NewPermissionDeniedError(msg string) PermissionDeniedError {
	return PermissionDeniedError{msg: msg}
}

type PermissionDeniedError struct {
	msg string
}

func (p PermissionDeniedError) Error() string {
	if p.msg == "" {
		return "permission denied"
	} else {
		return p.msg
	}
}
