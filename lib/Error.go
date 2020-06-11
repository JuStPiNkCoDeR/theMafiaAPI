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
	out += e.ParentError.Error() + "\n\n"
	return out + e.Message
}
