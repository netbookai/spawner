package common

func Int64Ptr(i int64) *int64 {
	return &i
}

func StrPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

func MapPtr(b map[string]string) *map[string]string {
	return &b
}
