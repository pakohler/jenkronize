package jenkins

type jenkinsError struct {
	context string
	err     error
}

func (e *jenkinsError) Error() string {
	if e.err != nil {
		return e.context + " - " + e.err.Error()
	} else {
		return e.context
	}
}

func newJenkinsError(ctx string, failure error) *jenkinsError {
	return &jenkinsError{
		context: ctx,
		err:     failure,
	}
}
