package search

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

var domainIndex map[string]int64
var reverseIndex map[string]int64

func loadIndex(indexFileName string, globalIndex *map[string]int64) error {
	indexFile, err := os.Open(indexFileName)
	if err != nil {
		return err
	}

	index := map[string]int64{}

	scanner := bufio.NewScanner(indexFile)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":::")
		if len(parts) < 2 {
			continue
		}
		key := parts[0]
		pos, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return err
		}

		index[key] = pos
	}

	*globalIndex = index

	return nil
}

func LoadDomainIndex(indexFileName string) error {
	return loadIndex(indexFileName, &domainIndex)
}

func LoadReverseIndex(indexFileName string) error {
	return loadIndex(indexFileName, &reverseIndex)
}
