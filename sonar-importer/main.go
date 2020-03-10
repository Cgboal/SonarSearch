package main

import (
	"bufio"
	"context"
	"github.com/Cgboal/DomainParser"
	"github.com/valyala/fastjson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/semaphore"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"github.com/schollz/progressbar"
)

type SonarResult struct {
	Timestamp   string `bson:"timestamp"`
	Name        string `bson:"name"`
	Type        string `bson:"type"`
	Value       string `bson:"value"`
	DomainIndex string `bson:"domain_index"`
}

func ParseFile(file *os.File, ch chan<- SonarResult, bar *progressbar.ProgressBar) {
	var wg sync.WaitGroup
	lock := semaphore.NewWeighted(int64(runtime.NumCPU()) / 2)
	scanner := bufio.NewScanner(file)
	domain_parser := parser.NewDomainParser()
	for scanner.Scan() {
		bar.Write([]byte(scanner.Text()))
		wg.Add(1)
		lock.Acquire(context.TODO(), 1)

		var p fastjson.Parser

		line := strings.TrimSpace(scanner.Text())
		go func(ch chan<- SonarResult) {
			defer wg.Done()
			defer lock.Release(1)

			var sonar_result SonarResult
			v, _ := p.Parse(line)
			sonar_result.Name = string(v.GetStringBytes("name"))
			sonar_result.Timestamp = string(v.GetStringBytes("timestamp"))
			sonar_result.Type = string(v.GetStringBytes("type"))
			sonar_result.Value = string(v.GetStringBytes("value"))
			sonar_result.DomainIndex = string(v.GetStringBytes("domain_index"))
			if sonar_result.DomainIndex == "" {
				sonar_result.DomainIndex = domain_parser.GetDomain(sonar_result.Name)
			}

			ch <- sonar_result
		}(ch)

	}
	wg.Wait()

	close(ch)
}

func Output(collection *mongo.Collection, ch <-chan SonarResult) {
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

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("sonar").Collection("A")



	subdomain_channel := make(chan SonarResult, 50000)

	go Output(collection, subdomain_channel)
	ParseFile(file, subdomain_channel, bar)

}
