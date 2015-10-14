package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMainHandler(t *testing.T) {
	rootRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal("Root request error: %s", err)
	}

	cases := []struct {
		w                    *httptest.ResponseRecorder
		r                    *http.Request
		accessTokenHeader    string
		expectedResponseCode int
		expectedResponseBody []byte
		expectedLogs         []string
	}{
		{
			w:                    httptest.NewRecorder(),
			r:                    rootRequest,
			accessTokenHeader:    "magic",
			expectedResponseCode: http.StatusOK,
			expectedResponseBody: []byte("You have some magic in you\n"),
			expectedLogs: []string{
				"Allowed an access attempt\n",
			},
		},
		{
			w:                    httptest.NewRecorder(),
			r:                    rootRequest,
			accessTokenHeader:    "",
			expectedResponseCode: http.StatusForbidden,
			expectedResponseBody: []byte("You don't have enough magic in you\n"),
			expectedLogs: []string{
				"Denied an access attempt\n",
			},
		},
	}

	for _, c := range cases {
		logReader, logWriter := io.Pipe()
		bufLogReader := bufio.NewReader(logReader)
		log.SetOutput(logWriter)

		c.r.Header.Set("X-Access-Token", c.accessTokenHeader)

		go func() {
			for _, expectedLine := range c.expectedLogs {
				msg, err := bufLogReader.ReadString('\n')
				if err != nil {
					t.Errorf("Expected to be able to read from log but got error: %s", err)
				}
				if !strings.HasSuffix(msg, expectedLine) {
					t.Errorf("Log line didn't match suffix:\n\t%q\n\t%q", expectedLine, msg)
				}
			}
		}()

		mainHandler(c.w, c.r)

		if c.expectedResponseCode != c.w.Code {
			t.Errorf("Status Code didn't match:\n\t%q\n\t%q", c.expectedResponseCode, c.w.Code)
		}

		if !bytes.Equal(c.expectedResponseBody, c.w.Body.Bytes()) {
			t.Errorf("Body didn't match:\n\t%q\n\t%q", string(c.expectedResponseBody), c.w.Body.String())
		}
	}
}
