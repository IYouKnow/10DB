package project

const (
	MaxProjectsPerUser     = 1
	MaxDatabasesPerProject = 3
)

type LimitError struct {
	message string
}

func NewLimitError(message string) *LimitError {
	return &LimitError{message: message}
}

func (e *LimitError) Error() string {
	return e.message
}
