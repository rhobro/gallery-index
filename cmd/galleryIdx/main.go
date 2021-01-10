package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Bytesimal/goutils/pkg/fileio"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)


var debugURL, _ = url.Parse("http://localhost:9090")
var cli = http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyURL(debugURL),
	},
}

const (
	Base = "https://wgsf.org.uk"
)

type Event struct {
	SRC    string    `json:"src"`
	DataID int       `json:"data-id"`
	Date   time.Time `json:"date"`

	Description string `json:"description"`
	Images []string `json:"images"`
}

func init() {
	flag.StringVar(&outPath, "o", "", "path to dir for output JSON file")
	flag.BoolVar(&formatJSON, "f", false, "controls if output JSON file is formatted")
	flag.DurationVar(&rateLim, "l", 1, "time between requests for rate-limiting")
	flag.Parse()

	// use tmp path as output if not specified by user
	if outPath == "" {
		fileio.Init("", "wgsf_gallery_idx_*")
		outPath = fileio.TmpDir
	}
	log.Printf("JSON output in %s", outPath)
}

var outPath string
var formatJSON bool
var rateLim time.Duration

func main() {
	var events []Event
	t := time.NewTicker(rateLim)

	var i int
	for {
		// rq
		rq, _ := http.NewRequest("GET", fmt.Sprintf("https://wgsf.org.uk/ajax/filter/gallery/%d", i), nil)
		rq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.101 Safari/537.36")
		rsp, err := cli.Do(rq)
		<-t.C
		if err != nil {
			log.Fatalf("can't request page: %s", err)
		}
		// add to raw HTML to make it parsable
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			log.Fatalf("can't parse html: %s", err)
		}
		sl := page.Find("article")
		rsp.Body.Close()
		if sl.Length() == 0 {
			break
		}

		sl.Each(func(_ int, sl *goquery.Selection) {
			// extract src
			src, _ := sl.Attr("class")
			src = strings.Split(src, " ")[0] // remove other css classes
			// extract data-id
			rawDataID, _ := sl.Attr("data-id")
			dataID, _ := strconv.Atoi(rawDataID)
			// extract description
			description := strings.TrimSpace(sl.Find("p.description").Get(0).FirstChild.Data)
			// date
			rawTime, _ := sl.Find("p > time").Attr("datetime")
			date, _ := time.Parse("2006-01-02T15:04:05-07:00", rawTime)

			// extract img urls
			var images []string
			sl.Each(func(_ int, sl *goquery.Selection) {
				rawImgHTML := sl.Find("a.image.lightbox.hidden")
				rawImgHTML.Each(func(_ int, sl *goquery.Selection) {
					path, ok := sl.Attr("href")
					if ok {
						u, _ := url.Parse(Base + path)
						images = append(images, Base + u.Query().Get("src"))
					}
				})
			})

			// add
			events = append(events, Event{
				SRC: src,
				DataID: dataID,
				Date: date,
				Description: description,
				Images: images,
			})
		})

		i += sl.Length()
	}

	// save to JSON file
	var bs []byte
	if formatJSON {
		bs, _ = json.MarshalIndent(&events, "", "    ")
	} else {
		bs, _ = json.Marshal(&events)
	}
	f, _ := os.Create(filepath.Join(outPath, "gallery_idx.json"))
	defer f.Close()
	_, _ = f.Write(bs)
}