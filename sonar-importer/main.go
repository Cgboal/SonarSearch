package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Cgboal/DomainParser"
	"github.com/schollz/progressbar"
	"github.com/valyala/fastjson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var output_wg sync.WaitGroup

type SonarResult struct {
	Timestamp string `bson:"timestamp"`
	Type      string `bson:"type"`
	Value     string `bson:"value"`
	Domain    string `bson:"domain"`
	Tld       string `bson:"tld"`
	Subdomain string `bson:"subdomain"`
}

func ParseFile(file *os.File, ch chan<- SonarResult, bar *progressbar.ProgressBar) {
	var wg sync.WaitGroup
	scanner := bufio.NewScanner(file)
	domain_parser := parser.NewDomainParser()

	linesChannel := make(chan string, 10000)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func(ch chan<- SonarResult) {
			defer wg.Done()
			var p fastjson.Parser
			for line := range linesChannel {
				var sonar_result SonarResult
				v, _ := p.Parse(line)
				name := string(v.GetStringBytes("name"))
				sonar_result.Timestamp = string(v.GetStringBytes("timestamp"))
				sonar_result.Type = string(v.GetStringBytes("type"))
				sonar_result.Value = string(v.GetStringBytes("value"))
				sonar_result.Domain = string(v.GetStringBytes("domain"))

				if sonar_result.Domain == "" {
					domain_parts := strings.Split(name, ".")
					offset := domain_parser.FindTldOffset(domain_parts)
					sonar_result.Domain = domain_parts[offset]
					sonar_result.Tld = strings.Join(domain_parts[offset+1:], ".")
					sonar_result.Subdomain = strings.Join(domain_parts[:offset], ".")
				}

				ch <- sonar_result
			}
		}(ch)
	}

	for scanner.Scan() {
		bar.Write([]byte(scanner.Text()))
		line := strings.TrimSpace(scanner.Text())
		linesChannel <- line

	}
	wg.Wait()

	close(ch)
}

func Output(collection *mongo.Collection, ch <-chan SonarResult) {
	defer output_wg.Done()
	docs := []interface{}{}
	for entry := range ch {
		docs = append(docs, entry)
		if len(docs) == 1000000 {
			ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
			opts := options.InsertMany().SetOrdered(false)
			_, err := collection.InsertMany(ctx, docs, opts)
			if err != nil {
				log.Println(err)
			}
			docs = []interface{}{}
		}
	}
}

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	filename := os.Args[2]
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	bar := progressbar.NewOptions64(
		fi.Size(),
		progressbar.OptionThrottle(20*time.Second),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	collection_name := os.Args[1]
	collection := client.Database("sonar").Collection(collection_name)

	subdomain_channel := make(chan SonarResult, 50000)

	output_wg.Add(1)
	go Output(collection, subdomain_channel)
	ParseFile(file, subdomain_channel, bar)
	output_wg.Wait()

}
