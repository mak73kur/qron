package qron

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Job struct {
	// Schedule parameters
	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string
	// Message body
	Payload []byte
}

// Check if job should be executed at this time
func (j Job) Match(now time.Time) bool {
	return cmp(j.Minute, now.Minute()) &&
		cmp(j.Hour, now.Hour()) &&
		cmp(j.Month, int(now.Month())) &&
		cmp(j.DayOfWeek, int(now.Weekday()))
}

// Compare the parameter expression with a given date property
func cmp(expr string, now int) bool {
	for _, v := range strings.Split(expr, ",") {
		if v == "*" || v == strconv.Itoa(now) {
			return true
		} else if len(v) > 2 && v[0:2] == "*/" {
			if div, _ := strconv.Atoi(v[2:]); div > 0 && now%div == 0 {
				return true
			}
		}
	}
	return false
}

// Parse multi-line tab string
func ParseTab(src string) ([]Job, error) {
	jobs := []Job{}
	for i, s := range strings.Split(src, "\n") {
		s = strings.TrimSpace(s)

		// skip comments and empty lines
		if s == "" || s[0:1] == "#" || s[0:2] == "//" {
			continue
		}

		job, err := ParseJob(s)
		if err != nil {
			return nil, fmt.Errorf("malformed tab on line %d, %s", i+1, err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// Create a job entity from a tab line
func ParseJob(src string) (Job, error) {
	job := Job{}
	vals := strings.SplitN(src, " ", 6)

	if len(vals) < 6 {
		return job, fmt.Errorf("wrong number of parameters: expected 6, got %d", len(vals))
	}

	for i, v := range vals {
		if i < 5 {
			match, _ := regexp.MatchString("^([0-9,/\\*])+$", v)
			if !match {
				return job, fmt.Errorf("invalid expression in %d parameter: '%s'", i+1, v)
			}
		}

		switch i {
		case 0:
			job.Minute = v
		case 1:
			job.Hour = v
		case 2:
			job.DayOfMonth = v
		case 3:
			job.Month = v
		case 4:
			job.DayOfWeek = v
		case 5:
			job.Payload = []byte(v)
		}
	}
	return job, nil
}
