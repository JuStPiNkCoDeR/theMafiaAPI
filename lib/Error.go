// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Cascading error implementation
package lib

import "fmt"

// Common stack error struct
type StackError struct {
	ParentError error  // Parent error which spawned this error
	Message     string // String contains short information about error
}

// Return complete error message
func (e *StackError) Error() string {
	var out = ""

	fmt.Println(e.ParentError)
	if e.ParentError != nil {
		out += e.ParentError.Error() + "\n\n\t"
	}

	return out + e.Message
}
