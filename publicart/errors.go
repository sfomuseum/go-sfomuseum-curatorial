package publicart

import (
	"fmt"
)

type NotFound struct{ code string }

func (e NotFound) Error() string {
	return fmt.Sprintf("Public art work '%s' not found", e.code)
}

func (e NotFound) String() string {
	return e.Error()
}

type MultipleCandidates struct{ code string }

func (e MultipleCandidates) Error() string {
	return fmt.Sprintf("Multiple candidates for public art work '%s'", e.code)
}

func (e MultipleCandidates) String() string {
	return e.Error()
}

func IsNotFound(e error) bool {

	switch e.(type) {
	case NotFound, *NotFound:
		return true
	default:
		return false
	}
}

func IsMultipleCandidates(e error) bool {

	switch e.(type) {
	case MultipleCandidates, *MultipleCandidates:
		return true
	default:
		return false
	}
}
