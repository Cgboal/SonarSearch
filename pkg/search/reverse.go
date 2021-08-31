package search

import (
	"bufio"
	"errors"
	"os"
	"strconv"

	"fmt"
	"io"
	"strings"

	"github.com/cgboal/sonarsearch/pkg/ipconv"
)

type ReverseSearch struct {
	file          *os.File
	needle        reverseNeedle
	reader        *bufio.Reader
	reverseResult reverseResult
	err           error
	foundFirst    bool
}

type reverseResult struct {
	Domain string
	IPv4   string
}
type reverseNeedle struct {
	Min uint32
	Max uint32
}

func newReverseNeedle(query string) (reverseNeedle, error) {
	if !strings.Contains(query, "/") {
		query = query + "/32"
	}
	min, max, err := ipconv.CIDRMinMaxInt(query)
	if err != nil {
		return reverseNeedle{}, err
	}

	needle := reverseNeedle{
		Min: min,
		Max: max,
	}

	return needle, nil
}

func NewReverseSearch(inputFileName string, query string) (*ReverseSearch, error) {
	if query == "" {
		return nil, errors.New("query cannot be blank")
	}

	needle, err := newReverseNeedle(query)
	if err != nil {
		return nil, err
	}

	needleIndex := ipconv.RoundDecIP(needle.Min, 1000)

	if err != nil {
		return nil, err
	}

	needleIndexString := fmt.Sprint(needleIndex)

	pos, err := getPos(needleIndexString)

	if err != nil {
		return nil, err
	}

	if pos == 0 {
		return nil, errors.New("no results found")
	}

	reader, file, err := getReader(inputFileName, pos)
	if err != nil {
		return nil, err
	}

	reverseSearch := ReverseSearch{
		file:       file,
		needle:     needle,
		reader:     reader,
		foundFirst: false,
	}

	return &reverseSearch, nil

}

func (rs *ReverseSearch) Next() bool {
	for {
		line, err := rs.reader.ReadBytes('\n')
		if err == io.EOF {
			rs.err = err
			return false
		}

		lineParts := strings.Split(string(line), ",")
		candidateIPv4, err := strconv.ParseUint(lineParts[0], 10, 32)
		candidateUInt32 := uint32(candidateIPv4)
		if err != nil {
			rs.err = err
		}

		if candidateUInt32 < rs.needle.Min || candidateUInt32 > rs.needle.Max {
			if rs.foundFirst {
				return false
			} else if candidateUInt32 > rs.needle.Max {
				return false
			} else {
				continue
			}
		}
		rs.foundFirst = true
		rs.reverseResult = reconstructReverseResult(candidateUInt32, lineParts[1])
		return true
	}
}

func (rs *ReverseSearch) Collect() map[string][]string {
	resultsMap := map[string][]string{}

	for rs.Next() {
		_, exists := resultsMap[rs.Result().IPv4]
		if !exists {
			resultsMap[rs.Result().IPv4] = []string{}
		}
		resultsMap[rs.Result().IPv4] = append(resultsMap[rs.Result().IPv4], rs.Result().Domain)

		if rs.err == io.EOF {
			break
		}
	}

	return resultsMap
}

func (rs *ReverseSearch) Result() reverseResult {
	return rs.reverseResult
}

func (rs *ReverseSearch) Close() {
	rs.file.Close()
}

func (rs *ReverseSearch) Error() error {
	return rs.err
}

func reconstructReverseResult(IPv4Uint uint32, domain string) reverseResult {
	ipv4String := ipconv.IntToIPv4(IPv4Uint)

	reverseResult := reverseResult{
		Domain: strings.TrimRight(domain, "\n"),
		IPv4:   ipv4String,
	}

	return reverseResult

}
