package main

import (
	"bufio"
	"flag"
	"fmt"
	parser "github.com/Cgboal/DomainParser"
	"github.com/cgboal/sonarsearch/pkg/ipconv"
	jsoniter "github.com/json-iterator/go"
	"log"
	"os"
	"runtime"
	"sync"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var dp parser.Parser

func init() {
	dp = parser.NewDomainParser()
}

type SonarEntry struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type FormatterFunc func(SonarEntry) (string, error)

func DomainLookupFormatter(entry SonarEntry) (string, error) {
	domainStruct := dp.ParseDomain(entry.Name)
	outputLine := fmt.Sprintf("%s,%s,%s\n", domainStruct.Domain, domainStruct.TLD, domainStruct.Subdomain)

	return outputLine, nil
}

func ReverseDomainLookupFormatter(entry SonarEntry) (string, error) {
	ipv4Int, err := ipconv.IPv4ToInt(entry.Value)
	if err != nil {
		return "", err
	}
	outputLine := fmt.Sprintf("%d,%s\n", ipv4Int, entry.Name)
	return outputLine, nil
}

func Map(formatterFunc FormatterFunc, inputChan <-chan []byte, lines chan string) {
	for line := range inputChan {
		var entry SonarEntry
		err := json.Unmarshal(line, &entry)
		if err != nil {
			//log.Println(err)
			fmt.Println(err)
			continue
		}

		if entry.Type != "a" {
			continue
		}

		outputLine, err := formatterFunc(entry)
		if err != nil {
			fmt.Println(err)
			continue
		}

		lines <- outputLine
	}
}

func Reducer(outputFileName string, lines <-chan string) {
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatal(err)
	}

	writer := bufio.NewWriter(outputFile)

	for line := range lines {
		writer.WriteString(line)
	}

	writer.Flush()
}

func main() {
	inputFileName := flag.String("i", "", "file path for raw sonar dataset")
	outputFileName := flag.String("o", "", "file path to store new, sorted dataset")
	format := flag.String("f", "", "what output format to use, can be 'domain' or 'reverse'")

	flag.Parse()

	if *inputFileName == "" || *outputFileName == "" || *format == "" {
		flag.Usage()
	}

	var formatter FormatterFunc
	if *format == "domain" {
		formatter = DomainLookupFormatter
	} else if *format == "reverse" {
		formatter = ReverseDomainLookupFormatter
	} else {
		fmt.Println("Format must be either 'domain' or 'reverse', got " + *format)
		os.Exit(1)
	}

	inputFile, err := os.Open(*inputFileName)
	if err != nil {
		log.Fatal(err)
	}

	var mapWg sync.WaitGroup
	var reduceWg sync.WaitGroup

	scanner := bufio.NewScanner(inputFile)
	scanner.Split(bufio.ScanLines)

	lines := make(chan string, 1024)

	reduceWg.Add(1)
	go func() {
		defer reduceWg.Done()
		Reducer(*outputFileName, lines)
	}()

	inputChan := make(chan []byte, 100000)

	for x := 0; x < runtime.NumCPU(); x++ {
		mapWg.Add(1)
		go func() {
			defer mapWg.Done()
			Map(formatter, inputChan, lines)
		}()
	}

	for scanner.Scan() {
		inputChan <- []byte(scanner.Text())
	}
	close(inputChan)

	mapWg.Wait()
	close(lines)
	reduceWg.Wait()

}
