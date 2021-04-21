package main

import (
	"encoding/json"
	"flag"
	"github.com/rhobro/goutils/pkg/fileio"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/wgsfGalleryIdx/internal/core"
	"github.com/rhobro/wgsfGalleryIdx/internal/idx"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func init() {
	wd, _ := os.Getwd()
	flag.StringVar(&jsonIn, "i", "", "json path to gallery index")
	flag.StringVar(&out, "o", wd, "directory in which to download images")
	flag.DurationVar(&rateLim, "l", time.Nanosecond, "rate limit duration if not using input json file")
	flag.Parse()

	if out == "" {
		fileio.Init("", "wgsf_gallery_dwld_*")
		out = fileio.TmpDir
	}
	log.Printf("images output in %s", out)
}

var (
	jsonIn  string
	out     string
	rateLim time.Duration
)

func main() {

	// load events
	var events []*core.Event
	if jsonIn == "" {
		events = idx.Index(rateLim)
	} else {
		bs, err := ioutil.ReadFile(jsonIn)
		if err != nil {
			log.Fatalf("can't read file at %s: %s", jsonIn, err)
		}
		err = json.Unmarshal(bs, &events)
		if err != nil {
			log.Fatalf("can't unmarshal json file at %s: %s", jsonIn, err)
		}
	}

	// dwld
	for _, e := range events {
		if len(e.Images) > 0 {
			// group into dirs with name as data-id
			root := filepath.Join(out, strconv.Itoa(e.DataID))
			err := os.MkdirAll(root, os.ModePerm)
			if err != nil {
				log.Fatalf("can't create dir at %s: %s", root, err)
			}

			// dwld and save imgs
			for i, img := range e.Images {
				imgPath := filepath.Join(root, strconv.Itoa(i)+filepath.Ext(img))
				out, err := os.Create(imgPath)
				if err != nil {
					log.Fatalf("can't create img file at %s: %s", imgPath, err)
				}

				// download
				rq, _ := http.NewRequest("GET", img, nil)
				rsp, err := httputil.RQUntil(http.DefaultClient, rq)
				if err != nil {
					log.Printf("can't rq img at %s: %s", img, err)
				}
				_, err = io.Copy(out, rsp.Body)
				if err != nil {
					log.Printf("can't copy file contents into file at %s: %s", imgPath, err)
				}
				rsp.Body.Close()

				log.Printf("Downloaded %5d : %s", e.DataID, filepath.Base(img))

				out.Close()
				err = os.Chmod(imgPath, os.ModePerm)
				if err != nil {
					log.Printf("can't change perms on img file at %s: %s", imgPath, err)
				}
			}
		}
	}
}
