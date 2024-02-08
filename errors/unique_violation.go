package errors

type UniqueViolationError struct {
	Message string
}

func (e *UniqueViolationError) Error() string {
	if e.Message == "" {
		return "unique constraint violation"
	}
	return e.Message
}

func (e *UniqueViolationError) Is(target error) bool {
	_, ok := target.(*UniqueViolationError)
	return ok
}

func NewUniqueViolationError() error {
	return &UniqueViolationError{}
}
