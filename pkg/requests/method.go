package requests

import "io"

func Get(url string, headers map[string]string, body io.Reader, timeout int) (Response, error) {
	return Request("GET", url, headers, body, timeout)
}

func Post(url string, headers map[string]string, body io.Reader, timeout int) (Response, error) {
	return Request("POST", url, headers, body, timeout)
}

func Put(url string, headers map[string]string, body io.Reader, timeout int) (Response, error) {
	return Request("PUT", url, headers, body, timeout)
}

func Delete(url string, headers map[string]string, body io.Reader, timeout int) (Response, error) {
	return Request("DELETE", url, headers, body, timeout)
}
