package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
)

const (
	defaultMaxWorkers         = "10"
	maxBodySize               = 1024 * 10
	parallelArg               = "-parallel"
	incorrectParallelErrorMsg = "Incorrect \"-parallel\" argument"
	insufficentArgsError      = "You should provide at least one URL"
	requestScheme             = "http"
)

type RequestError struct {
	StatusCode int
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("MyHTTP_Error: Status code %d", e.StatusCode)
}

type MyHTTPError struct {
	Message string
}

func (e *MyHTTPError) Error() string {
	return fmt.Sprintf("MyHTTP_Error: %s", e.Message)
}

type HashedURLResponse struct {
	url    string
	result string
}

func makeRequest(uri string, client http.Client) (string, error) {
	resp, err := client.Get(uri)
	if err != nil {
		return "", &MyHTTPError{err.Error()}
	}
	StatusCodeNotWithin2xx := (resp.StatusCode < http.StatusOK) || (resp.StatusCode > 299)
	if StatusCodeNotWithin2xx {
		return "", &RequestError{resp.StatusCode}
	}

	defer resp.Body.Close()
	if resp.ContentLength > maxBodySize {
		return "", &MyHTTPError{fmt.Sprintf("Response body above %d bytes threshold: %d", maxBodySize, resp.ContentLength)}
	}
	//Content-length might not be always available
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", &MyHTTPError{err.Error()}
	}

	return string(body), nil
}

func hashResponse(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

func makeRequests(urls []string, maxWorkers int) []HashedURLResponse {
	pool := make(chan int, maxWorkers)
	responses := []HashedURLResponse{}
	defer close(pool)

	wg := sync.WaitGroup{}
	var mutex sync.Mutex

	for _, url := range urls {
		client := http.Client{}
		pool <- 1
		wg.Add(1)
		go func(url string) {
			body, err := makeRequest(url, client)
			var msg string
			if err != nil {
				msg = err.Error()
			} else {
				msg = hashResponse(body)
			}
			mutex.Lock()
			responses = append(responses, HashedURLResponse{url, msg})
			mutex.Unlock()
			wg.Done()
			<-pool
		}(url)

	}

	wg.Wait()
	return responses
}

func parseURLs(urls []string) []string {
	parsedURLS := []string{}
	for _, uri := range urls {
		u, err := url.Parse(uri)
		if err != nil {
			// leave as it is, request will fail and user will be able to see
			parsedURLS = append(parsedURLS, uri)
			continue
		}
		if u.Scheme == "" {
			u.Scheme = "http"
		}
		parsedURLS = append(parsedURLS, u.String())

	}
	return parsedURLS

}

func parseArgs(args []string) (int, []string, error) {
	workers := 0

	var workerArgs []string
	urls := []string{}
	if len(args) == 0 {
		return workers, urls, errors.New(insufficentArgsError)
	}

	if (len(args) == 1) && (args[0] == parallelArg) {
		workerArgs = []string{parallelArg, "incorrect"}
		urls = []string{}
	} else {
		if args[0] == parallelArg {
			workerArgs = args[0:2]
			urls = args[2:]
		} else {
			workerArgs = []string{parallelArg, defaultMaxWorkers}
			urls = args
		}
	}

	workers, err := strconv.Atoi(workerArgs[1])
	if (err != nil) || workers < 1 {
		return 0, []string{}, errors.New(incorrectParallelErrorMsg)
	}

	urls = parseURLs(urls)

	return workers, urls, nil

}

func main() {
	args := os.Args[1:]
	workers, urls, err := parseArgs(args)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	results := makeRequests(urls, workers)

	for _, r := range results {
		fmt.Println(r.url, " ", r.result)
	}
	os.Exit(0)
}
