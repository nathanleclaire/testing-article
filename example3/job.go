package main

import (
	"log"
	"net/http"
	"time"
)

type Job interface {
	Log(...interface{})
	Suspend() error
	Resume() error
	Run() error
}

type PollerJob struct {
	suspend     chan bool
	resume      chan bool
	resourceUrl string
	inMemLog    string
}

func NewPollerJob(resourceUrl string) PollerJob {
	return PollerJob{
		resourceUrl: resourceUrl,
		suspend:     make(chan bool),
		resume:      make(chan bool),
	}
}

func (p PollerJob) Log(args ...interface{}) {
	log.Println(args...)
}

func (p PollerJob) Suspend() error {
	p.suspend <- true
	return nil
}

func (p PollerJob) PollServer() error {
	resp, err := http.Get(p.resourceUrl)
	if err != nil {
		return err
	}

	p.Log(p.resourceUrl, "--", resp.Status)

	return nil
}

func (p PollerJob) Run() error {
	for {
		select {
		case <-p.suspend:
			<-p.resume
		default:
			if err := p.PollServer(); err != nil {
				p.Log("Error trying to get resource: ", err)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (p PollerJob) Resume() error {
	p.resume <- true
	return nil
}

func main() {
	p := NewPollerJob("http://nathanleclaire.com")
	go p.Run()
	time.Sleep(5 * time.Second)

	p.Log("Suspending monitoring of server for 5 seconds...")
	p.Suspend()
	time.Sleep(5 * time.Second)

	p.Log("Resuming job...")
	p.Resume()

	// Wait for a bit before exiting
	time.Sleep(5 * time.Second)
}
