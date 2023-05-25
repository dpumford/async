package main

import (
	"async/pkg/function_queue"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	factQueue := function_queue.NewFunctionQueue[int, factResult](3)

	for i := 0; i < 10; i++ {
		factQueue.Run(getFact, i)
	}

	results := factQueue.Wait()

	for _, v := range results {
		if v.Err != nil {
			panic(v.Err)
		}

		fmt.Printf("Fact %d: %s\n", v.R.index, v.R.fact.Text)
	}
}

type factResult struct {
	index int
	fact  fact
}

type fact struct {
	Text string
}

func getFact(index int) (factResult, error) {
	res, err := http.Get("https://cat-fact.herokuapp.com/facts/random")

	if err != nil {
		return factResult{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return factResult{}, fmt.Errorf("received status code %d: %s", res.StatusCode, res.Status)
	}

	jsonReader := json.NewDecoder(res.Body)
	var f fact

	if decodeErr := jsonReader.Decode(&f); decodeErr != nil {
		return factResult{}, decodeErr
	}

	return factResult{
		index: index,
		fact:  f,
	}, nil
}
