package flags

type HelpError string

func (e HelpError) Error() string {
	return string(e)
}

type UsageError string

func (e UsageError) Error() string {
	return string(e)
}
