package merry

// Unwrap implements the go2 errors proposal.  Unwrapping a merry error returns the
// error's cause, same as merry.Cause(err).
func (e *merryErr) Unwrap() error {
	return e.Cause()
}

// merry errors have both a chain of wrapped errors (ending in a root error),
// as well as a chain of "causes".  merry's public API allows you to traverse
// the chain of causes, by recursively calling merry.Cause(error), and it allows
// you to get the root of the wrapped chain of merry errors, via merry.Unwrap(error).
// There was no public API for traversing the intermediary merryErrs in the chain
// of wrapped errors, but internal code in the package traverses it by recursively
// accessing merryErr.err.

// To meet the spirit of the new errors.Is and errors.As functions, those functions
// need to traverse both the chain of wrapped errors, as well as the chain of
// causes.  To achieve this, merryErr.Unwrap() returns the
// error's cause, and noCauseErr.Unwrap() returns the wrapped err.

// This implementation of merryErr.Is() and merryErr.As() casts the receiver
// as a noCauseErr, then recursively calls errors.Is/As on that.  This will compare
// the arg to each error in the chain, calling noCauseErr.Unwrap() at each step,
// which returns the wrapped error.  Once that recursion completes, if a match
// isn't found, the call unwinds back to the original errors.Is/As call, which
// is operating on the merryErr.  merryErr.Unwrap() is called, which returns
// the "cause" error (if any), and recurses on that error's Unwrap() implemention.
// In a large graph of merry errors with causes that are also merry errors, the
// recursion will traverse each nodes wrapped errors first, then traverse causes.
// In other words, it's a depth-first traversal, where wrapped errors are
// child nodes, and causes are sibling nodes.

type noCauseErr merryErr

func (u *noCauseErr) Unwrap() error {
	return u.err
}

func (u *noCauseErr) Error() string {
	return u.err.Error()
}

// Is implements the new go 2.0 errors function.  It returns true if
// the argument equals the current error, any of the errors in the merry
// wrapper chain, or the cause of the error.  It searches the chain
// of wrappers first, then tries the error's cause.
func (e *merryErr) Is(err error) bool {
	u := (*noCauseErr)(e)
	return is(u, err)
}

func (e *merryErr) As(target interface{}) bool {
	u := (*noCauseErr)(e)
	return as(u, target)
}
