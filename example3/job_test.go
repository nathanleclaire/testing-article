package main

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type ReadableLogger interface {
	Logger
	Read() string
}

type MessageReader struct {
	Msg string
}

func (mr *MessageReader) Read() string {
	return mr.Msg
}

type LastEntryLogger struct {
	*MessageReader
}

func (lel *LastEntryLogger) Log(args ...interface{}) {
	lel.Msg = fmt.Sprint(args...)
}

type DiscardFirstWriteLogger struct {
	*MessageReader
	writtenBefore bool
}

func (dfwl *DiscardFirstWriteLogger) Log(args ...interface{}) {
	if dfwl.writtenBefore {
		dfwl.Msg = fmt.Sprint(args...)
	}
	dfwl.writtenBefore = true
}

type FakeServerPoller struct {
	result string
	err    error
}

func (fsp FakeServerPoller) PollServer() (string, error) {
	return fsp.result, fsp.err
}

func TestPollerJobRunLog(t *testing.T) {
	defaultTestWaitInterval := 100 * time.Millisecond
	shortInterval := 20 * time.Millisecond
	longInterval := 200 * time.Millisecond

	testCases := []struct {
		expectedMsg string
		p           PollerJob
		logger      ReadableLogger
		sp          ServerPoller
	}{
		{
			expectedMsg: "200 OK",
			p:           NewPollerJob("madeup.website", shortInterval),
			logger:      &LastEntryLogger{&MessageReader{}},
			sp:          FakeServerPoller{"200 OK", nil},
		},
		{
			expectedMsg: "500 SERVER ERROR",
			p:           NewPollerJob("down.website", shortInterval),
			logger:      &LastEntryLogger{&MessageReader{}},
			sp:          FakeServerPoller{"500 SERVER ERROR", nil},
		},
		{
			expectedMsg: "Error trying to get state: DNS probe failed",
			p:           NewPollerJob("error.website", shortInterval),
			logger:      &LastEntryLogger{&MessageReader{}},
			sp:          FakeServerPoller{"", errors.New("DNS probe failed")},
		},
		{
			expectedMsg: "",
			p:           NewPollerJob("some.website", longInterval),

			// Discard first write since we want to verify that no
			// additional logs get made after the first one (time
			// out)
			logger: &DiscardFirstWriteLogger{MessageReader: &MessageReader{}},

			sp: FakeServerPoller{"200 OK", nil},
		},
	}

	for _, c := range testCases {
		c.p.Logger = c.logger
		c.p.ServerPoller = c.sp

		go c.p.Run()

		time.Sleep(defaultTestWaitInterval)

		if c.logger.Read() != c.expectedMsg {
			t.Errorf("Expected message did not align with what was written:\n\t%q != %q", c.expectedMsg, c.logger.Read())
		}
	}
}
