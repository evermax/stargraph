package update

// Updator is just a wrapper of string
type Updator string

// NewUpdator will create a new updator
func NewUpdator() Updator {
	return Updator("updator")
}

// Type will return the string "updator" to implement the interface.
func (c Updator) Type() string {
	return string(c)
}