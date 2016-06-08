package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var (
	url       = flag.String("url", "komikterbaru.com", "ini menjadi web target kita")
	start     = flag.Int("start", 1, "Dimulai dari chapter berapa?")
	chap      = flag.Int("chap", 1, "Berapa banyak chapter yang kita download ?")
	seperator = flag.String("sep", "chapter-", "ini yang akan membedakan satu chap denga yang lainnya")
)

var listURL []string
var downloadUrl string

func main() {
	localAddr := "127.9.90.1"
	localAddress, _ := net.ResolveTCPAddr("tcp", localAddr)
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			LocalAddr: localAddress,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
	}

	flag.Parse()

	for i := 0; i < *chap; i++ {
		if len(listURL) < 1 {
			downloadUrl = *url
		} else {
			downloadUrl = listURL[i+*start-1]
		}

		fmt.Println(downloadUrl)
		res, err := client.Get(downloadUrl)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			fmt.Println("Status Code not 200, get ", res.StatusCode)
			return
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		grabAllImage(string(body), *url)
		// getNextURL(string(body))
		if len(listURL) < 1 {
			getNextURL(string(body))
		}

	}
}

func getNextURL(content string) {
	// oldChap := fmt.Sprintf("%s0%d", *seperator, index)
	// nextChap := fmt.Sprintf("%s0%d", *seperator, index+1)
	// url := strings.Replace(*url, oldChap, nextChap, -1)

	parentStart := strings.Index(content, "<select class=\"select_list_chapter\" onchange=\"chapter_redirect(this.value);\">")
	parentEnd := strings.Index(content[parentStart:], "</option></select>")

	list := content[parentStart:parentEnd] + "</option>"
	doc, err := html.Parse(strings.NewReader(list))
	if err != nil {
		log.Fatal(err)
	}
	listChapterLink(doc)

}

func grabAllImage(content, url string) {
	parentStart := strings.Index(content, "<div class=\"chapter_images")
	parentEnd := strings.LastIndex(content[parentStart:], "</div>")

	figure := content[parentStart:parentEnd]
	doc, err := html.Parse(strings.NewReader(figure))
	if err != nil {
		log.Fatal(err)
	}
	parentDir := url[strings.LastIndex(url, "/"):strings.Index(url, *seperator)]
	nameDir := strings.Trim(url[strings.LastIndex(url, "/"):], ".html")
	dir := strings.Replace(parentDir+"/"+nameDir, "-", "_", -1)
	err = os.MkdirAll(strings.TrimLeft(dir, "/"), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	searchImage(doc, strings.TrimLeft(dir, "/"))

}

func searchImage(n *html.Node, nameDir string) {
	if n.Type == html.ElementNode && n.Data == "img" {
		for _, a := range n.Attr {
			if a.Key == "src" {
				downloadImage(a.Val, nameDir)
				fmt.Println(a.Val)
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		searchImage(c, nameDir)
	}
}

func listChapterLink(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "option" {
		for _, a := range n.Attr {
			if a.Key == "value" {
				listURL = append(listURL, a.Val)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		listChapterLink(c)
	}
}

func downloadImage(url, nameDir string) {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	fmt.Println(" GET ", url)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Println("Code not 200 in download image we get, ", res.StatusCode)
		panic("error download")
	}
	name := url[strings.LastIndex(url, "/")+1:]
	filename := fmt.Sprintf("%s/%s", nameDir, name)

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		panic(err)
	}
}
