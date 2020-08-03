// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Cascading error implementation
package lib

// Common stack error struct
type StackError struct {
	ParentError error  // Parent error which spawned this error
	Message     string // String contains short information about error
}

// Return complete error message
func (e *StackError) Error() string {
	var out = ""

	if e.ParentError != nil {
		out += e.ParentError.Error() + "\n\t"
	}

	return out + e.Message
}

// Make StackError
func Wrap(err error, message string) *StackError {
	return &StackError{
		ParentError: err,
		Message:     message,
	}
}