package main

import (
	"errors"
	"reflect"
	"testing"
)

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
	cases := []struct {
		f           FakeReleaseInfoer
		repo        string
		expectedMsg string
		expectedErr error
	}{
		{
			f: FakeReleaseInfoer{
				Tag: "v0.1.0",
				Err: nil,
			},
			repo:        "doesnt/matter",
			expectedMsg: "The latest release is v0.1.0",
			expectedErr: nil,
		},
		{
			f: FakeReleaseInfoer{
				Tag: "v0.1.0",
				Err: errors.New("TCP timeout"),
			},
			repo:        "doesnt/foo",
			expectedMsg: "",
			expectedErr: errors.New("Error querying GitHub API: TCP timeout"),
		},
	}

	for _, c := range cases {
		msg, err := getReleaseTagMessage(c.f, c.repo)
		if !reflect.DeepEqual(err, c.expectedErr) {
			t.Fatalf("Expected err to be %q but it was %q", c.expectedErr, err)
		}

		if c.expectedMsg != msg {
			t.Fatalf("Expected %q but got %q", c.expectedMsg, msg)
		}
	}
}
