package qron

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Job struct {
	// Schedule parameters
	Minute     []byte
	Hour       []byte
	DayOfMonth []byte
	Month      []byte
	DayOfWeek  []byte
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
func cmp(expr []byte, now int) bool {
	for _, v := range bytes.Split(expr, []byte(",")) {
		if bytes.Equal(v, []byte("*")) || bytes.Equal(v, []byte(strconv.Itoa(now))) {
			return true
		} else if len(v) > 2 && bytes.Equal(v[0:2], []byte("*/")) {
			if div, _ := strconv.Atoi(string(v[2:])); div > 0 && now%div == 0 {
				return true
			}
		}
	}
	return false
}

// Parse multi-line tab string
func ParseTab(src []byte) ([]Job, error) {
	jobs := []Job{}

	scanner := bufio.NewScanner(bytes.NewReader(src))
	for i := 1; scanner.Scan(); i++ {
		s := bytes.TrimSpace(scanner.Bytes())
		// skip comments and empty lines
		if len(s) == 0 || bytes.Equal(s[0:1], []byte("#")) || (len(s) > 1 && bytes.Equal(s[0:2], []byte("//"))) {
			continue
		}

		job, err := ParseJob(s)
		if err != nil {
			return nil, fmt.Errorf("malformed tab on line %d, %s", i, err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// Create a job entity from a tab line
func ParseJob(src []byte) (Job, error) {
	job := Job{}
	vals := bytes.SplitN(src, []byte(" "), 6)

	if len(vals) < 6 {
		return job, fmt.Errorf("wrong number of parameters: expected 6, got %d", len(vals))
	}

	for i, v := range vals {
		if i < 5 {
			match, _ := regexp.Match("^([0-9,/\\*])+$", v)
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
			job.Payload = v
		}
	}
	return job, nil
}
