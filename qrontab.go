package qron

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	Exp []string
	// Message body
	Payload string
	// Additional tags
	Tags map[string]interface{}
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
		writeLog(lvlDebug, fmt.Sprintf("parsed job: %+v %+v %s", job.Exp, job.Tags, job.Payload))
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// Create a job entity from a tab line
func ParseJob(src []byte) (Job, error) {
	job := Job{Exp: make([]string, 5)}
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
			tags, n, err := ParseTags(v)
			if err != nil {
				return job, err
			}
			if tags != nil {
				job.Tags = tags
				v = bytes.TrimRight(v[:len(v)-n], " ")
			}
			job.Payload = string(v)
		}
	}
	return job, nil
}

// Parse message options
func ParseTags(src []byte) (map[string]interface{}, int, error) {
	tagSep := []byte("`")
	if bytes.HasSuffix(src, tagSep) {
		start := bytes.LastIndex(src[:len(src)-1], tagSep)
		if start != -1 {
			var tags map[string]interface{}
			if err := json.Unmarshal(src[start+1:len(src)-1], &tags); err != nil {
				return nil, 0, err
			}
			return tags, len(src) - start, nil
		}
	}
	return nil, 0, nil
}
