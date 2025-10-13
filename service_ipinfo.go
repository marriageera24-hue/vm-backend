package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

func ipInfoBatchFetch(path string, items []string) map[int][]byte {
	batchSize := 100

	count := int(math.Ceil(float64(len(items)) / float64(batchSize)))

	batches := make(map[int][]string, count)
	max := len(items)

	i := 0
	for i < count {
		from := i * batchSize
		to := from + batchSize

		if to > max {
			to = max
		}

		batches[i] = items[from:to]
		i++
	}

	url := "https://ipinfo.io/batch?token=" + os.Getenv("IPINFO_TOKEN")

	responses := make(map[int][]byte, len(batches))

	if len(items) == 0 {
		return responses
	}

	for i, batch := range batches {
		client := &http.Client{
			Timeout: time.Second * 30,
		}

		items := make([]string, len(batch))
		for i, item := range batch {
			items[i] = strings.Replace(path, ":value", item, 1)
		}

		b, err := json.Marshal(items)
		if err != nil {
			continue
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
		if err != nil {
			continue
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		responses[i] = body
	}

	return responses
}
