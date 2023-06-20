package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestUnmarshalOk(t *testing.T) {
	jsonString, err := os.ReadFile("test/data/okSample1.json")
	if err != nil {
		t.Fatal(err)
	}
	want := Data{
		Houses: []House{
			{ID: 1,
				Address:   "4 Pumpkin Hill Street Antioch, TN 37013",
				Homeowner: "Nicole Bone",
				Price:     105124,
				PhotoURL:  "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg",
			},
		},
		Ok: true,
	}
	got, err := unmarshal(jsonString)
	if got.Ok != want.Ok {
		t.Errorf("got %v, wanted %v", got.Ok, want.Ok)
	}
	if got.Message != "" {
		t.Errorf("got %s, wanted empty", got.Message)
	}
	if len(got.Houses) != len(want.Houses) {
		t.Errorf("houses size got %d, wanted %d", len(got.Houses), len(want.Houses))
	}
	var gotHouse House = got.Houses[0]
	var wantHouse House = want.Houses[0]

	if gotHouse.Address != wantHouse.Address {
		t.Errorf("got %s, wanted %s", gotHouse.Address, wantHouse.Address)
	}
	if gotHouse.Homeowner != wantHouse.Homeowner {
		t.Errorf("got %s, wanted %s", gotHouse.Homeowner, wantHouse.Homeowner)
	}
	if gotHouse.PhotoURL != wantHouse.PhotoURL {
		t.Errorf("got %s, wanted %s", gotHouse.PhotoURL, wantHouse.PhotoURL)
	}
	if gotHouse.Price != wantHouse.Price {
		t.Errorf("got %d, wanted %d", gotHouse.Price, wantHouse.Price)
	}

}

func TestUnmarshalNotOk(t *testing.T) {
	jsonString, err := os.ReadFile("test/data/notOkSample1.json")
	if err != nil {
		t.Fatal(err)
	}
	want := Data{
		Ok:      false,
		Message: "Service Unavailable",
	}
	got, err := unmarshal(jsonString)
	if got.Ok != want.Ok {
		t.Errorf("got %v, wanted %v", got.Ok, want.Ok)
	}
	if got.Message != want.Message {
		t.Errorf("got %s, wanted %s", got.Message, want.Message)
	}
	if len(got.Houses) != len(want.Houses) {
		t.Errorf("houses size got %d, wanted %d", len(got.Houses), len(want.Houses))
	}
}

func TestBuildUrl(t *testing.T) {
	want := "http://app-homevision-staging.herokuapp.com/api_project/houses?page=1&per_page=20"
	got := buildUrl(1, 20)
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestBuildPath(t *testing.T) {
	want := outputFolder + "/1-4 Pumpkin Hill Street Antioch, TN 37013.jpg"
	house := House{
		ID:        1,
		Address:   "4 Pumpkin Hill Street Antioch, TN 37013",
		Homeowner: "Nicole Bone",
		Price:     105124,
		PhotoURL:  "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg",
	}
	got := buildPath(house)
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

// http client tests

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestGetPageErrorHandling(t *testing.T) {

	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		equals(t, req.URL.String(), "http://app-homevision-staging.herokuapp.com/api_project/houses?page=1&per_page=10")
		return &http.Response{
			StatusCode: 503,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`{"message":"Service Unavailable","ok":false}`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	api := API{client}
	isLastPage, err := api.getPage(1)
	equals(t, false, isLastPage)
	if err == nil {
		t.Errorf("expecting error")
	}
	equals(t, errors.New("API returned unexpected response status =  in page = 1"), err)
}

func TestGetPageErrorHandlingWhenInconsistentJson(t *testing.T) {

	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		equals(t, req.URL.String(), "http://app-homevision-staging.herokuapp.com/api_project/houses?page=1&per_page=10")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`{"houses":[{"],"ok":true}`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	api := API{client}
	isLastPage, err := api.getPage(1)
	equals(t, false, isLastPage)
	if err == nil {
		t.Errorf("expecting error")
	}
	equals(t, errors.New("API returned inconsistent json in page = 1"), err)
}

func TestGetPageLastPageWhenCurrentPageReturnsLessThanRequested(t *testing.T) {

	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		equals(t, req.URL.String(), "http://app-homevision-staging.herokuapp.com/api_project/houses?page=1&per_page=10")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`{"houses":[{"id":0,"address":"4 Pumpkin Hill Street Antioch, TN 37013","homeowner":"Nicole Bone","price":105124,"photoURL":"https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg"},{"id":1,"address":"495 Marsh Road Portage, IN 46368","homeowner":"Rheanna Walsh","price":161856,"photoURL":"https://media-cdn.tripadvisor.com/media/photo-s/09/7c/a2/1f/patagonia-hostel.jpg"},{"id":2,"address":"7088 N. Wild Rose Ave. Hartford, CT 06106","homeowner":"Maurice Sparrow","price":219714,"photoURL":"https://images.adsttc.com/media/images/5e5e/da62/6ee6/7e7b/b200/00e2/medium_jpg/_fi.jpg"},{"id":3,"address":"52 South Ridge St. Vienna, VA 22180","homeowner":"Lucca Benson","price":152639,"photoURL":"https://image.shutterstock.com/image-photo/traditional-english-semidetached-house-260nw-231369511.jpg"},{"id":4,"address":"7798 Poplar St. Stillwater, MN 55082","homeowner":"Adelle Steadman","price":222178,"photoURL":"https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg"},{"id":5,"address":"606 Silver Spear Lane Defiance, OH 43512","homeowner":"Haya Pena","price":236163,"photoURL":"https://image.shutterstock.com/image-photo/houses-built-circa-1960-on-260nw-177959672.jpg"},{"id":6,"address":"9590 8th Lane Seymour, IN 47274","homeowner":"Kimora Redfern","price":265730,"photoURL":"https://i.pinimg.com/originals/47/b9/7e/47b97e62ef6f28ea4ae2861e01def86c.jpg"},{"id":7,"address":"362 Lancaster Dr. Oak Forest, IL 60452","homeowner":"Waqar Lister","price":191706,"photoURL":"https://i.pinimg.com/originals/47/b9/7e/47b97e62ef6f28ea4ae2861e01def86c.jpg"},{"id":8,"address":"9230 W. Howard Street Warminster, PA 18974","homeowner":"Kalum Vasquez","price":86738,"photoURL":"https://image.shutterstock.com/image-photo/houses-built-circa-1960-on-260nw-177959672.jpg"}],"ok":true}`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	api := API{client}
	isLastPage, err := api.getPage(1)
	equals(t, true, isLastPage)
	if err != nil {
		t.Errorf("unexpected error ")
	}
}

func TestGetPageLastPageWhenCurrentPageReturnsEmpty(t *testing.T) {

	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		equals(t, req.URL.String(), "http://app-homevision-staging.herokuapp.com/api_project/houses?page=1&per_page=10")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`{"houses":[],"ok":true}`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	api := API{client}
	isLastPage, err := api.getPage(1)
	equals(t, true, isLastPage)
	if err != nil {
		t.Errorf("unexpected error ")
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, expected, actual)
		tb.FailNow()
	}
}
