package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

var wg sync.WaitGroup

func main() {
	//get the link
	var link string
	if len(os.Args) < 2 {
		fmt.Println("no arguments")
		return
	}
	link = os.Args[1]
	urlCheck(link)

	bname, tid, tr, imgs, b := handleLink(link)
	var pb = progressbar.NewOptions(
		int(imgs),
		progressbar.OptionShowElapsedTimeOnFinish(),
		// progressbar.OptionOnCompletion(func() {
		// 	// fmt.Println(config.elapsed)
		// 	OptionShowElapsedTimeOnFinish()
		// 	//wait for commi
		// }),
	)

	wg.Add(int(imgs)) //add the total no of routines that are going to run to the watchsyncgroup

	var i int64 = 0
	for i <= tr {
		if b.Posts[i].Tim == 0 {
			i++
			continue
		}
		go b.GetFile(i, bname, tid, pb)
		i++
	}

	wg.Wait() //wait for it
}

func lol() {

}

func handleLink(link string) (string, string, int64, int64, board) {
	re1 := regexp.MustCompile(`boards.4chan.org|boards.4channel.org`)
	link = re1.ReplaceAllString(link, "a.4cdn.org")

	re2 := regexp.MustCompile(`a.4cdn.org/(.+?)/thread/([0-9]*)`)
	m := re2.FindStringSubmatch(link)
	bname := "/" + m[1] + "/"
	tid := m[2]

	//mkdir returns err
	e := os.Mkdir(tid, fs.ModePerm)
	if e != nil && !os.IsExist(e) {
		fmt.Println("directory already exists")
		log.Fatal(e)
	}

	resp, err := http.Get(link + ".json")
	if err != nil {
		fmt.Println("fetch error")
		log.Fatal(err)
	}
	defer resp.Body.Close()

	b_json := &board{}
	json.NewDecoder(resp.Body).Decode(b_json)
	tr := b_json.Posts[0].Replies
	imgs := b_json.Posts[0].Images

	return bname, tid, tr, imgs, *b_json
}

func (b board) GetFile(i int64, bname string, tid string, pb *progressbar.ProgressBar) {
	defer wg.Done() //decrement syncgroup count when the job is done

	fname := b.Posts[i].Filename
	tim := b.Posts[i].Tim
	ext := b.Posts[i].Ext
	fp := tid + "/" + fname + ext

	file, err := os.Create(fp)

	if err != nil {
		fmt.Println("error while creating file")
		log.Fatal(err)
	}
	defer file.Close()

	url := "https://i.4cdn.org" + bname + strconv.FormatInt(tim, 10) + ext
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Cant get the file " + fname)
		log.Fatal(err)
	}
	defer resp.Body.Close()

	_, e := io.Copy(file, resp.Body)
	if e != nil {
		fmt.Println("error while copying file contents")
		log.Fatal(e)
	}

	pb.Add(1)
}

func urlCheck(link string) {
	m, _ := regexp.MatchString(`^https://boards.4chan(nel)*.org/.+?/thread/\d*$`, link)
	if !m {
		fmt.Println("wrong url")
		os.Exit(1)
	}
}
