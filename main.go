package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	apiEndpoint = "https://api.openai.com/v1/chat/completions"
	filterWords = "tourism,travel,destination,hotel,flight"
)

var historyLog []string

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/ask", ask)
	http.HandleFunc("/history", showHistory)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func ask(w http.ResponseWriter, r *http.Request) {
	question := r.FormValue("question")
	apiKey := "sk-proj-u9Lq1hT8p57rkF2yhQVIT3BlbkFJ6siiX3FTC7t4csbdI0jS"
	client := resty.New()

	if !isQuestionValid(question) {
		fmt.Fprintf(w, "Your request was declined because your question is not related to the vision of the touristic company.")
		return
	}

	response, err := client.R().
		SetAuthToken(apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"model":      "gpt-3.5-turbo",
			"messages":   []interface{}{map[string]interface{}{"role": "system", "content": question}},
			"max_tokens": 50,
		}).
		Post(apiEndpoint)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error while sending the request: %v", err), http.StatusInternalServerError)
		return
	}

	body := response.Body()

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while decoding JSON response: %v", err), http.StatusInternalServerError)
		return
	}

	content := data["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	historyLog = append(historyLog, content)
	fmt.Fprintf(w, content)
}

func showHistory(w http.ResponseWriter, r *http.Request) {
	for _, item := range historyLog {
		fmt.Fprintf(w, "<p>%s</p>", item)
	}
}

func isQuestionValid(question string) bool {
	for _, word := range strings.Split(filterWords, ",") {
		if strings.Contains(strings.ToLower(question), word) {
			return true
		}
	}
	return false
}
