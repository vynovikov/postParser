package repo

// Min returns minimun number
func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Max returns maximum number
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
