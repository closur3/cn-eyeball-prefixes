package iputil

func First(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func AppendUnique(values []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		if value != "" && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}

func Clean(v []string) []string { return AppendUnique(nil, v...) }
