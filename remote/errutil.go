package main

type textError string

func (e textError) Error() string { return string(e) }

func newError(msg string) error {
	return textError(msg)
}
