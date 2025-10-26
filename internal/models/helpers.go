package models

// StringPtr returns a pointer to the string value passed in
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the int value passed in
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the bool value passed in
func BoolPtr(b bool) *bool {
	return &b
}
