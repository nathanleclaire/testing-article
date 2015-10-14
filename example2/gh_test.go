package main

import "testing"

type FakeReleaseInfoer struct {
	Tag string
	Err error
}

func (f FakeReleaseInfoer) GetLatestReleaseTag(repo string) (string, error) {
	if f.Err != nil {
		return "", f.Err
	}

	return f.Tag, nil
}

func TestGetReleaseTagMessage(t *testing.T) {
	f := FakeReleaseInfoer{
		Tag: "v0.1.0",
		Err: nil,
	}

	expectedMsg := "The latest release is v0.1.0"
	msg, err := getReleaseTagMessage(f, "dev/null")
	if err != nil {
		t.Fatalf("Expected err to be nil but it was %s", err)
	}

	if expectedMsg != msg {
		t.Fatalf("Expected %s but got %s", expectedMsg, msg)
	}
}
