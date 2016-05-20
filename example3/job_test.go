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
	waitBeforeReading := 100 * time.Millisecond
	shortInterval := 20 * time.Millisecond
	longInterval := 200 * time.Millisecond

	testCases := []struct {
		p           PollerJob
		logger      ReadableLogger
		sp          ServerPoller
		expectedMsg string
	}{
		{
			p:           NewPollerJob("madeup.website", shortInterval),
			logger:      &LastEntryLogger{&MessageReader{}},
			sp:          FakeServerPoller{"200 OK", nil},
			expectedMsg: "200 OK",
		},
		{
			p:           NewPollerJob("down.website", shortInterval),
			logger:      &LastEntryLogger{&MessageReader{}},
			sp:          FakeServerPoller{"500 SERVER ERROR", nil},
			expectedMsg: "500 SERVER ERROR",
		},
		{
			p:           NewPollerJob("error.website", shortInterval),
			logger:      &LastEntryLogger{&MessageReader{}},
			sp:          FakeServerPoller{"", errors.New("DNS probe failed")},
			expectedMsg: "Error trying to get state: DNS probe failed",
		},
		{
			p: NewPollerJob("some.website", longInterval),

			// Discard first write since we want to verify that no
			// additional logs get made after the first one (time
			// out)
			logger: &DiscardFirstWriteLogger{MessageReader: &MessageReader{}},

			sp:          FakeServerPoller{"200 OK", nil},
			expectedMsg: "",
		},
	}

	for _, c := range testCases {
		c.p.Logger = c.logger
		c.p.ServerPoller = c.sp

		go c.p.Run()

		time.Sleep(waitBeforeReading)

		if c.logger.Read() != c.expectedMsg {
			t.Errorf("Expected message did not align with what was written:\n\texpected: %q\n\tactual: %q", c.expectedMsg, c.logger.Read())
		}
	}
}

func TestPollerJobSuspendResume(t *testing.T) {
	p := NewPollerJob("foobar.com", 20*time.Millisecond)
	waitBeforeReading := 100 * time.Millisecond
	expectedLogLine := "200 OK"
	normalServerPoller := &FakeServerPoller{expectedLogLine, nil}

	logger := &LastEntryLogger{&MessageReader{}}
	p.Logger = logger
	p.ServerPoller = normalServerPoller

	// First start the job / polling
	go p.Run()

	time.Sleep(waitBeforeReading)

	if logger.Read() != expectedLogLine {
		t.Errorf("Line read from logger does not match what was expected:\n\texpected: %q\n\tactual: %q", expectedLogLine, logger.Read())
	}

	// Then suspend the job
	if err := p.Suspend(); err != nil {
		t.Errorf("Expected suspend error to be nil but got %q", err)
	}

	// Fake the log line to detect if poller is still running
	newExpectedLogLine := "500 Internal Server Error"
	logger.MessageReader.Msg = newExpectedLogLine

	// Give it a second to poll if it's going to poll
	time.Sleep(waitBeforeReading)

	// If this log writes, we know we are polling the server when we're not
	// supposed to (job should be suspended).
	if logger.Read() != newExpectedLogLine {
		t.Errorf("Line read from logger does not match what was expected:\n\texpected: %q\n\tactual: %q", newExpectedLogLine, logger.Read())
	}

	if err := p.Resume(); err != nil {
		t.Errorf("Expected resume error to be nil but got %q", err)
	}

	// Give it a second to poll if it's going to poll
	time.Sleep(waitBeforeReading)

	if logger.Read() != expectedLogLine {
		t.Errorf("Line read from logger does not match what was expected:\n\texpected: %q\n\tactual: %q", expectedLogLine, logger.Read())
	}
}
