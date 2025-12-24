package helpers

// StringValue safely dereferences a string pointer, returning an empty string if the pointer is nil.
func StringValue(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// Ptr returns a pointer to the given string.
func Ptr(s string) *string {
	return &s
}
