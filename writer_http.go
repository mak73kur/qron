package qron

import (
	"bytes"
	"fmt"
	"net/http"
	"io/ioutil"
)

type HTTPWriter struct {
	URL     string
	Method  string
	Headers map[string]string
}

func (w HTTPWriter) Write(msgBody []byte, tags map[string]interface{}) error {
	url, method, headers := w.URL, w.Method, w.Headers

	if tu, ok := tags["url"]; ok {
		if su, ok := tu.(string); ok && su != "" {
			url = su
		}
	}
	if tm, ok := tags["method"]; ok {
		if sm, ok := tm.(string); ok && sm != "" {
			method = sm
		}
	}
	if th, ok := tags["headers"]; ok {
		if mh, ok := th.(map[string]string); ok && len(mh) > 0 {
			for k, v := range mh {
				headers[k] = v
			}
		}
	}

	writeLog(lvlDebug, "making http request to " + url)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(msgBody))
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	httpClient := &http.Client{}
	res, err := httpClient.Do(req)

	if verb >= lvlDebug {
		respBody, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		writeLog(lvlDebug, fmt.Sprintf("http response %d, %s", res.StatusCode, respBody))
	}

	if res.StatusCode >= 400 {
		return fmt.Errorf("failed to make http request, response code is %d", res.StatusCode)
	}
	return err
}
