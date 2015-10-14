package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Logger interface {
	Log(...interface{})
}

type SuspendResumer interface {
	Suspend() error
	Resume() error
}

type Job interface {
	Logger
	SuspendResumer
	Run() error
}

type ServerPoller interface {
	PollServer() (string, error)
}

type PollerLogger struct{}

type URLServerPoller struct {
	resourceUrl string
}

type PollSuspendResumer struct {
	SuspendCh chan bool
	ResumeCh  chan bool
}

type PollerJob struct {
	WaitDuration time.Duration
	ServerPoller
	Logger
	*PollSuspendResumer
}

func NewPollerJob(resourceUrl string, waitDuration time.Duration) PollerJob {
	return PollerJob{
		WaitDuration: waitDuration,
		Logger:       &PollerLogger{},
		ServerPoller: &URLServerPoller{
			resourceUrl: resourceUrl,
		},
		PollSuspendResumer: &PollSuspendResumer{
			SuspendCh: make(chan bool),
			ResumeCh:  make(chan bool),
		},
	}
}

func (l *PollerLogger) Log(args ...interface{}) {
	log.Println(args...)
}

func (usp *URLServerPoller) PollServer() (string, error) {
	resp, err := http.Get(usp.resourceUrl)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(usp.resourceUrl, " -- ", resp.Status), nil
}

func (ssr *PollSuspendResumer) Suspend() error {
	ssr.SuspendCh <- true
	return nil
}

func (ssr *PollSuspendResumer) Resume() error {
	ssr.ResumeCh <- true
	return nil
}

func (p PollerJob) Run() error {
	for {
		select {
		case <-p.PollSuspendResumer.SuspendCh:
			<-p.PollSuspendResumer.ResumeCh
		default:
			state, err := p.PollServer()
			if err != nil {
				p.Log("Error trying to get state: ", err)
			} else {
				p.Log(state)
			}

			time.Sleep(p.WaitDuration)
		}
	}

	return nil
}

func main() {
	var j Job
	j = NewPollerJob("http://nathanleclaire.com", 1*time.Second)
	go j.Run()
	time.Sleep(5 * time.Second)

	j.Log("Suspending monitoring of server for 5 seconds...")
	j.Suspend()
	time.Sleep(5 * time.Second)

	j.Log("Resuming job...")
	j.Resume()

	// Wait for a bit before exiting
	time.Sleep(5 * time.Second)
}
