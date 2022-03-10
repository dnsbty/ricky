package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"
)

type Article struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type Query struct {
	Articles []Article `json:"random"`
}

type Result struct {
	Query Query `json:"query"`
}

func fetchPage(pageSize int) *Result {
	url := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=query&list=random&rnnamespace=0&rnlimit=%d&format=json", pageSize)

	client := http.Client{Timeout: time.Second * 5}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error creating request to the Wikipedia API: %+v", err)
		os.Exit(2)
	}

	req.Header.Set("User-Agent", "ricky-cli")

	res, getErr := client.Do(req)
	if getErr != nil {
		fmt.Println("Error fetching data from the Wikipedia API: %+v", getErr)
		os.Exit(2)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println("Error reading response from the Wikipedia API: %+v", readErr)
		os.Exit(2)
	}

	result := Result{}

	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		fmt.Println("Error parsing response from the Wikipedia API: %+v", jsonErr)
		os.Exit(2)
	}

	return &result
}

func displayArticle(article *Article) {
	articleUrl := fmt.Sprintf("https://en.wikipedia.org/?curid=%d", article.Id)
	openErr := browser.OpenURL(articleUrl)
	if openErr != nil {
		fmt.Println(article.Title)
	}
}

func findMatch(articles []Article, startChar string) *Article {
	for _, a := range articles {
		if strings.ToLower(startChar) == strings.ToLower(a.Title[0:1]) {
			return &a
		}
	}

	return nil
}

func main() {
	allowedAttempts := 10
	pageSize := 1
	startChar := ""
	numArgs := len(os.Args)

	if numArgs > 2 {

		fmt.Println("Only a single argument may be provided.")
		os.Exit(1)
	}

	if numArgs == 2 {
		pageSize = 100
		startChar = os.Args[1]

		if len(startChar) > 1 {
			fmt.Println("Only the starting character should be provided")
			os.Exit(1)
		}
	}

	if len(startChar) > 0 {
		remainingAttempts := allowedAttempts
		for remainingAttempts > 0 {
			result := fetchPage(pageSize)

			article := findMatch(result.Query.Articles, startChar)
			if article != nil {
				displayArticle(article)
				os.Exit(0)
			}
			remainingAttempts = remainingAttempts - 1
		}
		totalArticles := pageSize * allowedAttempts
		fmt.Println("Checked %d articles, but found none that matched started with %s. Try again if you'd like.", totalArticles)
		os.Exit(1)
	} else {
		result := fetchPage(pageSize)
		displayArticle(&result.Query.Articles[0])
		os.Exit(0)
	}
}
