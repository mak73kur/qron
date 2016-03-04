package qron

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	minute = iota
	hour
	dayOfMonth
	month
	dayOfWeek
)

type Job struct {
	// Schedule parameters
	Exp map[int]string
	// Message body
	Payload []byte
}

// Check if job should be executed at this time
func (j Job) Match(now time.Time) bool {
	return cmp(j.Exp[minute], now.Minute()) &&
		cmp(j.Exp[hour], now.Hour()) &&
		cmp(j.Exp[dayOfMonth], now.Day()) &&
		cmp(j.Exp[month], int(now.Month())) &&
		cmp(j.Exp[dayOfWeek], int(now.Weekday()))
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
	job := Job{Exp: map[int]string{}}
	vals := bytes.SplitN(src, []byte(" "), 6)

	if len(vals) < 6 {
		return job, fmt.Errorf("wrong number of parameters: expected 6, got %d", len(vals))
	}

	for i, v := range vals {
		switch {
		case i < 5:
			match, _ := regexp.Match("^([0-9,/\\*])+$", v)
			if !match {
				return job, fmt.Errorf("invalid expression in %d parameter: '%s'", i+1, v)
			}
			job.Exp[i] = string(v)
		case i == 5:
			job.Payload = v
		}
	}
	return job, nil
}
