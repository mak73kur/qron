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
	r Reader
	w Writer

	sync.Mutex
	tab []Job
}

func NewSchedule(r Reader, w Writer) *Schedule {
	return &Schedule{r: r, w: w}
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

	// reader can implement its own watch routine
	upd := make(chan []byte)
	if watcher, ok := sch.r.(Watcher); ok {
		go watcher.Watch(upd)
	}
	go func() {
		for {
			select {
			case <-sig:
				if err := sch.load(); err != nil {
					log.Println(err)
				}
			case tab := <-upd:
				if tab, err := ParseTab(tab); err == nil {
					sch.SetTab(tab)
				} else {
					log.Println(err)
				}
			}
		}
	}()
	return nil
}

func (sch *Schedule) load() error {
	b, err := sch.r.Read()
	if err != nil {
		return err
	}
	tab, err := ParseTab(b)
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
