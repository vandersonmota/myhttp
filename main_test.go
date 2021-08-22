package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
)

const (
	testURL = "http://google.com"
)

func assertEqual(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		t.Logf("Expected: %v - Actual: %v", expected, actual)
		t.Fail()
	}
}

func assertNotEqual(t *testing.T, expected, actual interface{}) {
	if expected == actual {
		t.Logf("Expected %v to not be equal to %v", expected, actual)
		t.Fail()
	}
}

func assertSliceEqual(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Logf("Expected: %v - Actual: %v", expected, actual)
		t.Fail()
	}
}

func TestMakeRequestOK(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte("responsebody"))
	}))
	defer testServer.Close()

	body, err := makeRequest(testServer.URL, *testServer.Client())
	assertEqual(t, nil, err)
	assertEqual(t, body, "responsebody")
}
func TestMakeRequestNotFound(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(404)
		res.Write([]byte("Not Found"))
	}))
	defer testServer.Close()

	body, err := makeRequest(testServer.URL, *testServer.Client())

	requestError := err.(*RequestError)
	if !errors.As(err, &requestError) {
		t.Log("UnexpectedError", err)
		t.Fail()
	}
	assertEqual(t, "MyHTTP_Error: Status code 404", err.Error())
	assertNotEqual(t, " ", body)
}
func TestMakeRequestServerError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(500)
		res.Write([]byte("SERVER ERROR"))
	}))
	defer testServer.Close()

	body, err := makeRequest(testServer.URL, *testServer.Client())

	requestError := err.(*RequestError)
	if !errors.As(err, &requestError) {
		t.Log("UnexpectedError", err)
		t.Fail()
	}
	if err.Error() != "MyHTTP_Error: Status code 500" {
		t.Log("Unexpected Exception Message", err)
		t.Fail()
	}
	if body != "" {
		t.Log("Unexpected body: ", body)
		t.Fail()
	}
}

func TestParseArgs(t *testing.T) {

	workers, urls, err := parseArgs([]string{"-parallel", "foobar"})
	assertEqual(t, 0, workers)
	assertSliceEqual(t, []string{}, urls)
	assertNotEqual(t, nil, err)
	assertEqual(t, "Incorrect \"-parallel\" argument", err.Error())

	workers, urls, err = parseArgs([]string{"-parallel"})
	assertEqual(t, 0, workers)
	assertSliceEqual(t, []string{}, urls)
	assertNotEqual(t, nil, err)
	assertEqual(t, "Incorrect \"-parallel\" argument", err.Error())

	workers, urls, err = parseArgs([]string{})
	assertEqual(t, 0, workers)
	assertSliceEqual(t, []string{}, urls)
	assertNotEqual(t, nil, err)
	assertEqual(t, "You should provide at least one URL", err.Error())

	workers, urls, err = parseArgs([]string{"google.com"})
	assertEqual(t, 10, workers)
	assertSliceEqual(t, []string{"http://google.com"}, urls)
	assertEqual(t, nil, err)

	workers, urls, err = parseArgs([]string{"-parallel", "5", "https://google.com", "example.com"})
	assertEqual(t, 5, workers)
	assertSliceEqual(t, []string{"https://google.com", "http://example.com"}, urls)
	assertEqual(t, nil, err)

	workers, urls, err = parseArgs([]string{"-parallel", "google.com", "apple.com"})
	assertEqual(t, 0, workers)
	assertSliceEqual(t, []string{}, urls)
	assertNotEqual(t, nil, err)
	assertEqual(t, "Incorrect \"-parallel\" argument", err.Error())

	workers, urls, err = parseArgs([]string{"-parallel", "-5", "google.com", "example.com"})
	assertEqual(t, 0, workers)
	assertSliceEqual(t, []string{}, urls)
	assertNotEqual(t, nil, err)
	assertEqual(t, "Incorrect \"-parallel\" argument", err.Error())

	workers, urls, err = parseArgs([]string{"-parallel", "0", "google.com", "example.com"})
	assertEqual(t, 0, workers)
	assertSliceEqual(t, []string{}, urls)
	assertNotEqual(t, nil, err)
	assertEqual(t, "Incorrect \"-parallel\" argument", err.Error())
}

func TestMakeRequests(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte("responsebody"))
	}))
	testServer2 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(500)
		res.Write([]byte("ServerError"))
	}))
	defer testServer.Close()
	defer testServer2.Close()

	expected := []HashedURLResponse{
		HashedURLResponse{testServer.URL, "029f0e2f5b6c5e1bc52e145415d95f8c"},
		HashedURLResponse{testServer2.URL, "MyHTTP_Error: Status code 500"},
	}
	sort.Slice(expected[:], func(i, j int) bool {
		return expected[i].result < expected[j].result
	})

	urls := []string{testServer.URL, testServer2.URL}
	hashedResults := makeRequests(urls, 1)

	sort.Slice(hashedResults[:], func(i, j int) bool {
		return hashedResults[i].result < hashedResults[j].result
	})

	assertSliceEqual(t, expected, hashedResults)
}
