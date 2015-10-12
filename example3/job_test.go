package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/log"
)

type WriteOnceLogger struct {
	msg string
}

func (wol *WriteOnceLogger) Log(args ...interface{}) {
	log.Warn(args)
	wol.msg = fmt.Sprint(args...)
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
	_ = 20 * time.Millisecond
	longInterval := 200 * time.Millisecond

	testCases := []struct {
		expectedMsg string
		p           PollerJob
		sp          ServerPoller
	}{
		/*
			{
				expectedMsg: "200 OK",
				p:           NewPollerJob("madeup.website", shortInterval),
				sp:          FakeServerPoller{"200 OK", nil},
			},
			{
				expectedMsg: "500 SERVER ERROR",
				p:           NewPollerJob("down.website", shortInterval),
				sp:          FakeServerPoller{"500 SERVER ERROR", nil},
			},
			{
				expectedMsg: "Error trying to get state: DNS probe failed",
				p:           NewPollerJob("error.website", shortInterval),
				sp:          FakeServerPoller{"", errors.New("DNS probe failed")},
			},
		*/
		{
			expectedMsg: "",
			p:           NewPollerJob("some.website", longInterval),
			sp:          FakeServerPoller{"200 OK", nil},
		},
	}

	for _, c := range testCases {
		logger := &WriteOnceLogger{}
		c.p.Logger = logger
		c.p.ServerPoller = c.sp

		go c.p.Run()

		time.Sleep(defaultTestWaitInterval)

		if logger.msg != c.expectedMsg {
			t.Errorf("Expected message did not align with what was written:\n\t%q != %q", c.expectedMsg, logger.msg)
		}
	}
}
