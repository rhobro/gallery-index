package main

import (
	"encoding/json"
	"flag"
	"github.com/rhobro/goutils/pkg/fileio"
	"github.com/rhobro/wgsfGalleryIdx/internal/idx"
	"log"
	"os"
	"path/filepath"
	"time"
)

func init() {
	wd, _ := os.Getwd()
	flag.StringVar(&outPath, "o", wd, "path to dir for output JSON file")
	flag.DurationVar(&rateLim, "l", time.Nanosecond, "time between requests for rate-limiting")
	flag.Parse()

	// use tmp path as output if not specified by user
	if outPath == "" {
		fileio.Init("", "wgsf_gallery_idx_*")
		outPath = fileio.TmpDir
	}
	log.Printf("JSON output in %s", outPath)
}

var outPath string
var rateLim time.Duration

func main() {
	events := idx.Index(rateLim)

	// save to JSON file
	var bs []byte
	bs, _ = json.MarshalIndent(&events, "", "    ")
	f, _ := os.Create(filepath.Join(outPath, "gallery_idx.json"))
	defer f.Close()
	_, _ = f.Write(bs)
}
