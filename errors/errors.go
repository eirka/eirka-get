package errors

var (
	ErrInvalidParam  = &RequestError{ErrorString: "Invalid Parameter", ErrorCode: 400}
	ErrInternalError = &RequestError{ErrorString: "Internal Error", ErrorCode: 500}
	ErrNotFound      = &RequestError{ErrorString: "Request Not Found", ErrorCode: 404}
)

type RequestError struct {
	ErrorString string
	ErrorCode   int
}

func (err *RequestError) Code() int {
	return err.ErrorCode
}

func (err *RequestError) Error() string {
	return err.ErrorString
}

func ErrorMessage(error_type *RequestError, args ...map[string]interface{}) (code int, message map[string]interface{}) {
	code = error_type.Code()
	message = map[string]interface{}{"error_message": error_type.Error()}

	return
}
