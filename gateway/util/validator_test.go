package util

import "testing"

func TestIsUrlValid(t *testing.T) {
	passUrl := "https://google.com"
	failUrl := "google.com"

	if !IsUrlValid(passUrl) {
		t.Errorf("Url %s should be valid", passUrl)
	}

	if IsUrlValid(failUrl) {
		t.Errorf("Url %s should be invalid", failUrl)
	}
}

func TestIsMatchRegex(t *testing.T) {
	pattern := `^(http|https):\/\/[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(\/\S*)?$`
	passUrl := "https://google.com"
	failUrl := "google.com"

	if !IsMatchRegex(pattern, passUrl) {
		t.Errorf("Url %s should be valid", passUrl)
	}

	if IsMatchRegex(pattern, failUrl) {
		t.Errorf("Url %s should be invalid", failUrl)
	}
}
