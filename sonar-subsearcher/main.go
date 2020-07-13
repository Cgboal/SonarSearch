package main

import (
	"bufio"
	"github.com/Cgboal/DomainParser"
	"github.com/valyala/fastjson"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"github.com/schollz/progressbar"
	"sync/atomic"
	"fmt"
)

type SonarResult struct {
	Timestamp   string `bson:"timestamp"`
	Name        string `bson:"name"`
	Type        string `bson:"type"`
	Value       string `bson:"value"`
	DomainIndex string `bson:"domain_index"`
}

func ParseFile(file *os.File, bar *progressbar.ProgressBar) {
	var wg sync.WaitGroup
	scanner := bufio.NewScanner(file)
	domain_parser := parser.NewDomainParser()

	var remote uint64
	var remotegw uint64
	var gw uint64
	var gateway uint64
	var citrix uint64
	var vpn uint64

	candidates := map[string]*uint64{
		"remote": &remote,
		"remotegw": &remotegw,
		"gw": &gw,
		"gateway": &gateway,
		"citrix": &citrix,
		"vpn": &vpn,
	}

	for scanner.Scan() {
		bar.Write([]byte(scanner.Text()))
		wg.Add(1)

		var p fastjson.Parser

		line := strings.TrimSpace(scanner.Text())
		go func() {
			defer wg.Done()
			v, _ := p.Parse(line)
			name := string(v.GetStringBytes("name"))
			subdomain := domain_parser.GetSubdomain(name)
			subdomain_parts := strings.Split(subdomain, ".")

			for _, part := range subdomain_parts {
				if _, ok := candidates[part]; ok {
					atomic.AddUint64(candidates[part], 1)
					break
				}
			}
		}()

	}
	wg.Wait()

	for k := range candidates {
		fmt.Printf("%s: %d\n", k, *candidates[k])
	}

}


func main() {
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	bar := progressbar.NewOptions(
		int(fi.Size()),
		progressbar.OptionThrottle(20*time.Second),
		progressbar.OptionSetBytes(int(fi.Size())),
	)

	ParseFile(file, bar)

}
