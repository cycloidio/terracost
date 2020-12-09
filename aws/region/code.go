package region

// Code represents an AWS region code.
type Code string

// NewFromZone returns the region code of the given zone or empty string if invalid.
func NewFromZone(zone string) Code {
	if len(zone) < 1 {
		return ""
	}
	return Code(zone[:len(zone)-1])
}

// NewFromName returns the region code from its name or empty string if invalid.
func NewFromName(name string) Code {
	return nameToCode[name]
}

// Valid returns true if the region exists and is supported, false otherwise.
func (c Code) Valid() bool {
	if c == "" {
		return false
	}
	_, ok := codeToName[c]
	return ok
}

// String returns the code of the region as a string.
func (c Code) String() string {
	return string(c)
}
