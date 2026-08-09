package cli

// ExitError signals that a command already printed everything the user
// needs to see (e.g. a prereq report) and main() just needs to exit with
// Code — no additional "Error: ..." line should be printed on top of it.
type ExitError struct{ Code int }

func (e *ExitError) Error() string { return "" }
