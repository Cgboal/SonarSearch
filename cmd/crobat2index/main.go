package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/cgboal/sonarsearch/pkg/ipconv"
)

type KeyFunc func(line string) string 

func getReader(fileName string) (*bufio.Reader, error) {
	inputFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(inputFile)
	return reader, nil
}

func domainKey(entry string) string {
	return entry
}

func reverseKey(entry string) string {
	entryInt, _ := strconv.ParseUint(entry, 10, 32)

	key := ipconv.RoundDecIP(uint32(entryInt), 10)
	return fmt.Sprintf("%d", key)
}

func generateIndex(keyFunc KeyFunc, inputFileName string) error {
	reader, err := getReader(inputFileName)
	if err != nil {
		return err
	}

	pos := int64(0)
	currentKey := ""

	for {
		line, err := reader.ReadBytes('\n')

		delimPos := bytes.IndexByte(line, ',')
		if delimPos != -1 {
			entry := string(line[:delimPos])
			key := keyFunc(entry)
			if key != currentKey {
				posString := fmt.Sprint(pos)
				fmt.Printf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(posString), posString)
				currentKey = key
			}
		}

		pos = pos + int64(len(line))

		if err == io.EOF {
			break
		}
	}

	return nil

}

func main() {
	inputFileName := flag.String("i", "", "file path for raw sonar dataset")
	format := flag.String("f", "", "what output format to use, can be 'domain' or 'reverse'")

	flag.Parse()

	if *inputFileName == "" || *format == "" {
		flag.Usage()
	}

	var keyFunc KeyFunc
	if *format == "domain" {
		keyFunc = domainKey
	} else if *format == "reverse" {
		keyFunc = reverseKey
	} else {
		fmt.Println("Format must be either 'domain' or 'reverse', got " + *format)
		os.Exit(1)
	}
	generateIndex(keyFunc, *inputFileName)
}
