package util

import "testing"

func TestGetMapperUrl(t *testing.T) {
	mapperUrl := GetMapperUrl()
	if mapperUrl != "http://mapper:1111" {
		t.Errorf("Mapper url is not correct: %s", mapperUrl)
	}
}

func TestGetRedirectUrl(t *testing.T) {
	redirectUrl := GetRedirectUrl()
	if redirectUrl != "http://redirect:2222" {
		t.Errorf("Redirect url is not correct", redirectUrl)
	}
}

func TestGenUUID(t *testing.T) {
	uuid := GenUUID()
	if uuid == "" {
		t.Errorf("UUID is empty")
	}

	uuid2 := GenUUID()
	if uuid == uuid2 {
		t.Errorf("UUID is duplicated")
	}
}
