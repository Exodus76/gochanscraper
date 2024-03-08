package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
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

var filenames []string

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
	fmt.Println("Board: ", boardName, "\nThread: ", threadId, "\nReplies: ", totalReplies, "\nImages: ", totalImgs)

	var wg sync.WaitGroup
	errChan := make(chan error)

	var pb = progressbar.NewOptions(int(totalImgs), progressbar.OptionShowElapsedTimeOnFinish())

	for i := range totalReplies {
		if newBoard.Posts[i].Tim != 0 {
			wg.Add(1)
			go func(i int64) {
				defer wg.Done()
				err := newBoard.GetPostImage(i, boardName, threadId, pb)
				if err != nil {
					errChan <- err
				}
			}(i)
		}
	}

	go func() {
		for err := range errChan {
			if err != nil {
				log.Fatalf("error while downloading image: %v", err)
			}
		}
	}()

	wg.Wait() //wait for it
	close(errChan)

	select {
	case <-errChan:
		fmt.Println("\nGenerating html...")
		generateHtml(threadId, threadLink, filenames)
	}
}

func (b *board) GetPostImage(index int64, boardName string, threadId string, pb *progressbar.ProgressBar) error {
	fileName := b.Posts[index].Filename
	tim := b.Posts[index].Tim
	extension := b.Posts[index].Ext

	filePath := path.Join(threadId, fileName+extension)

	filenames = append(filenames, fileName+extension)

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("error while creating file: %v", err)
	}
	defer file.Close()

	resp, err := http.Get(fmt.Sprintf("https://i.4cdn.org%s/%d%s", boardName, tim, extension))
	if err != nil {
		log.Fatalf("error while fetching image: %v", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatalf("error while writing to file %s: %v", fileName, err)
	}

	pb.Add(1)

	return nil
}

func parseLink(threadLink string, newBoard *board) (string, string, int64, int64, *board) {
	cdnLink := regexp.MustCompile(`boards.4chan.org|boards.4channel.org`).ReplaceAllString(threadLink, "a.4cdn.org")

	cdnLinkFinal := regexp.MustCompile(`a.4cdn.org/(.+?)/thread/([0-9]*)`).FindStringSubmatch(cdnLink)
	boardName := "/" + cdnLinkFinal[1] + "/"
	threadId := cdnLinkFinal[2]

	e := os.Mkdir(threadId, fs.ModePerm)
	if e != nil && !os.IsExist(e) {
		log.Fatalf("directory already exists: %v", e)
	}

	resp, err := http.Get(threadLink + ".json")
	if err != nil {
		log.Fatalf("fetch error: %v", err)
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

// helper function
func isImageFile(filename string) bool {
	lowercaseFilename := strings.ToLower(filename)
	return strings.HasSuffix(lowercaseFilename, ".jpg") ||
		strings.HasSuffix(lowercaseFilename, ".png") ||
		strings.HasSuffix(lowercaseFilename, ".gif")
}

func generateHtml(threadId string, threadLink string, fileNames []string) {
	f, err := os.Create("./" + threadId + "/_index.html")
	if err != nil {
		fmt.Println("error while creating html file")
		log.Fatal(err)
	}
	defer f.Close()

	t, err := template.New("template.html").Funcs(template.FuncMap{
		"isImageFile": isImageFile,
	}).ParseFiles("template.html")
	if err != nil {
		log.Fatalf("error while parsing template: %v", err)
	}

	data := struct {
		Title     string
		FileNames []string
	}{
		Title:     threadLink,
		FileNames: fileNames,
	}

	err = t.Execute(f, data)

	if err != nil {
		fmt.Println("error while writing to html file")
		log.Fatal(err)
	}
}
