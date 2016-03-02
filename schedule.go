package qron

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Schedule struct {
	l Loader
	w Writer

	sync.Mutex
	tab []Job
}

func NewSchedule(l Loader, w Writer) *Schedule {
	return &Schedule{l: l, w: w}
}

func (sch *Schedule) Tab() []Job {
	sch.Lock()
	defer sch.Unlock()
	return sch.tab
}

func (sch *Schedule) SetTab(new []Job) {
	sch.Lock()
	sch.tab = new
	sch.Unlock()
}

// Loads current schedule and starts a new routine to reload schedule on any changes
func (sch *Schedule) LoadAndWatch() error {
	if err := sch.load(); err != nil {
		return err
	}
	// register the signal channel
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGHUP)

	// loader can implement its own watch routine
	poll := make(chan string)
	pollErr := make(chan error)
	if poller, ok := sch.l.(Poller); ok {
		go poller.Poll(poll, pollErr)
	}

	go func() {
		for {
			select {
			case <-sig:
				if err := sch.load(); err != nil {
					log.Println(err)
				}
			case str := <-poll:
				if tab, err := ParseTab(str); err == nil {
					sch.SetTab(tab)
				}
			case err := <-pollErr:
				log.Println("Tab polling error:", err)
			}
		}
	}()
	return nil
}

func (sch *Schedule) load() error {
	str, err := sch.l.Load()
	if err != nil {
		return err
	}
	tab, err := ParseTab(str)
	if err != nil {
		return err

	}
	sch.SetTab(tab)
	return nil
}

// Runs the scheduler
func (sch *Schedule) Run() {
	// Wait for the beginning of the next minute
	time.Sleep(time.Now().Truncate(time.Minute).Add(time.Minute).Sub(time.Now()))

	// Schedule first run and start the ticker routine
	go sch.iterate(time.Now().UTC())
	for now := range time.Tick(time.Minute) {
		go sch.iterate(now.UTC())
	}
}

func (sch *Schedule) iterate(now time.Time) {
	for _, job := range sch.Tab() {
		if job.Match(now) {
			if err := sch.w.Write(job.Payload); err != nil {
				log.Println(err)
			}
		}
	}
}
