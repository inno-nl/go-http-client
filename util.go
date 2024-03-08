package httpclient

func getLengthOfLongestString(stringSlice []string) int {
	longestString := 0

	for _, s := range stringSlice {
		strLen := len(s)

		if strLen > longestString {
			longestString = strLen
		}
	}

	return longestString
}
