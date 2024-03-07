package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
)

type board struct {
	Posts []struct {
		Bumplimit   int64  `json:"bumplimit"`
		Com         string `json:"com"`
		Ext         string `json:"ext"`
		Filename    string `json:"filename"`
		Fsize       int64  `json:"fsize"`
		H           int64  `json:"h"`
		Imagelimit  int64  `json:"imagelimit"`
		Images      int64  `json:"images"`
		Md5         string `json:"md5"`
		Name        string `json:"name"`
		No          int64  `json:"no"`
		Now         string `json:"now"`
		Replies     int64  `json:"replies"`
		Resto       int64  `json:"resto"`
		SemanticURL string `json:"semantic_url"`
		Sub         string `json:"sub"`
		Tim         int64  `json:"tim"`
		Time        int64  `json:"time"`
		TnH         int64  `json:"tn_h"`
		TnW         int64  `json:"tn_w"`
		UniqueIps   int64  `json:"unique_ips"`
		W           int64  `json:"w"`
	} `json:"posts"`
}

func main() {
	//handle cli args
	var board board
	var threadLink string

	if len(os.Args) > 1 {
		threadLink = os.Args[1]
	} else {
		fmt.Println("Please provide a thread link")
		return
	}

	urlCheck(threadLink)

	// board := parseLink(threadLink, &newBoard)
	boardName, threadId, totalReplies, totalImgs, newBoard := parseLink(threadLink, &board)

	for i := range totalImgs {

		fmt.Println("Image: ", newBoard.Posts[i].Filename, "Size: ", newBoard.Posts[i].Fsize, "MD5: ", newBoard.Posts[i].Md5)
	}

	fmt.Println("Board: ", boardName, "Thread: ", threadId, "Replies: ", totalReplies, "Images: ", totalImgs, "Posts: ", len(newBoard.Posts))

}

func parseLink(threadLink string, newBoard *board) (string, string, int64, int64, *board) {
	cdnLink := regexp.MustCompile(`boards.4chan.org|boards.4channel.org`).ReplaceAllString(threadLink, "a.4cdn.org")

	cdnLinkFinal := regexp.MustCompile(`a.4cdn.org/(.+?)/thread/([0-9]*)`).FindStringSubmatch(cdnLink)
	boardName := "/" + cdnLinkFinal[1] + "/"
	threadId := cdnLinkFinal[2]

	e := os.Mkdir(threadId, fs.ModePerm)
	if e != nil && !os.IsExist(e) {
		log.Fatalf("directory already exists", e)
	}

	resp, err := http.Get(threadLink + ".json")
	if err != nil {
		log.Fatalf("fetch error", err)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&newBoard)

	totalReplies := newBoard.Posts[0].Replies
	totalImgs := newBoard.Posts[0].Images

	return boardName, threadId, totalReplies, totalImgs, newBoard
}

func urlCheck(threadLink string) {
	m := regexp.MustCompile(`^https://boards.4chan(nel)*.org/[^/]+/thread/\d*$`).Match([]byte(threadLink))

	if !m {
		log.Fatal("Invalid URL")
	}
}
