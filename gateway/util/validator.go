package util

import "regexp"

// Check if the provided url is valid, can be http or https and have a valid format
func IsUrlValid(url string) bool {
	pattern := `^(http|https):\/\/[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(\/\S*)?$`
	return IsMatchRegex(pattern, url)
}

// Check if the provided url is valid, can be http or https
func IsMatchRegex(pattern string, url string) bool {
	regexp, _ := regexp.Compile(pattern)
	return regexp.MatchString(url)
}
