package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// TODO move constants to configuration file
const outputFolder string = "output"
const totalPages int = 10
const pageSize int = 10
const retries int = 20
const retryDelay int = 100
const fileNameComponentSeparator string = "-"
const enableConcurrentOptimization bool = true // false can help testing and debugging
const channelSize = 50                         // relevant when enableConcurrentOptimization is true
const totalWorkers = 5                         // relevant when enableConcurrentOptimization is true

var houseChannel = make(chan House, channelSize) // TODO monitor channel size

var api = NewAPI()

type Data struct {
	Houses  []House `json:"houses"`
	Message string  `json:"message"`
	Ok      bool    `json:"ok"`
}

type House struct {
	ID        int    `json:"id"`
	Address   string `json:"address"`
	Homeowner string `json:"homeowner"`
	Price     int    `json:"price"`
	PhotoURL  string `json:"photoURL"`
}

func main() {
	var page int = 1
	var wg sync.WaitGroup
	for w := 1; w <= totalWorkers; w++ {
		wg.Add(1)
		go launchWorker(w, houseChannel, &wg)
	}
	fmt.Println("process starts")
	for {
		fmt.Println("Processing page ", page)
		isLastPage, err := getPageRetryable(page)
		if err != nil {
			log.Fatal(err)
		}
		if isLastPage {
			break // success
		}
		page++
		if page > totalPages {
			break
		}
	}
	fmt.Println("process step 1 completed: fetch completed")
	close(houseChannel)
	wg.Wait()
	fmt.Println("process step 2 completed: images downloaded")
	os.Exit(0)
}

func launchWorker(w int, houseChannel chan House, wg *sync.WaitGroup) {
	defer wg.Done()

	for house := range houseChannel {
		processWorker(w, house)
	}
}

func processWorker(w int, house House) {
	log.Println("Worker", w, "started  download ID=", house.ID)
	download(house)
	log.Println("Worker", w, "finished download ID=", house.ID)
}

func downloadAsync(house House) {
	houseChannel <- house
}

type API struct {
	Client *http.Client
}

func NewAPI() API {
	return API{Client: &http.Client{}}
}

func getPageRetryable(page int) (bool, error) {
	var isLastPage bool = false
	var err error
	for attempt := 1; attempt <= retries; attempt++ {
		isLastPage, err = api.getPage(page)
		if err == nil {
			return isLastPage, nil
		} else {
			var sleepMillis = (attempt - 1) * retryDelay
			time.Sleep(time.Duration(sleepMillis) * time.Millisecond)
		}
	}
	log.Println("failing after all retries, consider increasing number of retries, current value = " + strconv.Itoa(retries))
	return false, err
}

func buildUrl(page int, perPage int) string {
	return "http://app-homevision-staging.herokuapp.com/api_project/houses?page=" + strconv.Itoa(page) + "&per_page=" + strconv.Itoa(perPage)
}

func (api *API) getPage(page int) (bool, error) {
	resp, getErr := api.Client.Get(buildUrl(page, pageSize))
	if getErr != nil {
		log.Println(getErr)
		return false, getErr
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Println(readErr)
		return false, getErr
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("API returned unexpected response status = " + resp.Status + " in page = " + strconv.Itoa(page))
		return false, errors.New("API returned unexpected response status = " + resp.Status + " in page = " + strconv.Itoa(page))
	}
	data, jsonErr := unmarshal(body)
	if jsonErr != nil {
		log.Println(jsonErr)
		return false, errors.New("API returned inconsistent json in page = " + strconv.Itoa(page))
	}
	if !data.Ok {
		log.Println("API returned not ok response in page = " + strconv.Itoa(page))
		return false, errors.New("API returned not ok response in page = " + strconv.Itoa(page))
	}
	for _, house := range data.Houses {
		if enableConcurrentOptimization {
			downloadAsync(house)
		} else {
			download(house)
		}
	}
	var isLastPage = len(data.Houses) < pageSize || len(data.Houses) == 0
	return isLastPage, nil
}

func unmarshal(body []byte) (Data, error) {
	var data Data
	jsonErr := json.Unmarshal([]byte(body), &data)
	return data, jsonErr
}

func buildPath(house House) string { // TODO should we replace file spaces introduced by house address ?
	var ext = filepath.Ext(house.PhotoURL)
	return filepath.Join(outputFolder, strconv.Itoa(house.ID)+fileNameComponentSeparator+house.Address+ext)
}

func download(house House) {
	file, err := os.Create(buildPath(house))

	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	res, err := api.Client.Get(house.PhotoURL) // TODO consider retryable logic / library

	if err != nil {
		log.Println(err)
	}

	defer res.Body.Close()
	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Println(err)
	}
}
