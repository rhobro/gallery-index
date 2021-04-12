package idx

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/Bytesimal/wgsfGalleryIdx/internal/core"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const base = "https://wgsf.org.uk"

func init() {
	// use tor
	p, _ := url.Parse("socks5://localhost:9050")
	http.DefaultTransport = &http.Transport{
		Proxy: http.ProxyURL(p),
		MaxIdleConns: 1,
	}
}

func Index(rateLim time.Duration) []core.Event {
	var events []core.Event
	t := time.NewTicker(rateLim)

	var i int
	for {
		// rq
		rq, _ := http.NewRequest("GET", fmt.Sprintf("https://wgsf.org.uk/ajax/filter/gallery/%d", i), nil)
		rq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.101 Safari/537.36")
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
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
						u, _ := url.Parse(base + path)
						images = append(images, base+u.Query().Get("src"))
					}
				})
			})

			// add
			log.Printf("Event ID: %d", dataID)
			events = append(events, core.Event{
				SRC:         src,
				DataID:      dataID,
				Date:        date,
				Description: description,
				Images:      images,
			})
		})

		i += sl.Length()
	}

	return events
}
