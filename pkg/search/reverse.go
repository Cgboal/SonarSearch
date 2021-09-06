package search

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"time"

	"fmt"
	"io"
	"strings"

	"bytes"

	"github.com/cgboal/sonarsearch/pkg/ipconv"
	"github.com/spf13/viper"
)

type ReverseSearch struct {
	file          *os.File
	needle        reverseNeedle
	scanner       *bufio.Scanner
	reverseResult reverseResult
	err           error
	query string
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

type ReverseResponse struct {
	Results map[string][]string
	Err     error
}

type ReverseQuery struct {
	Query           string
	Take            int
	Skip            int
	ResponseChannel chan ReverseResponse
}

func NewReversePool(requests <-chan ReverseQuery) {
	for i := 0; i < 5; i++ {
		go startReverseWorker(requests)
	}
}

func startReverseWorker(requests <-chan ReverseQuery) {
	for query := range requests {
		searcher, err := NewReverseSearch(viper.GetString("reverse_file"), query.Query)
		if err != nil {
			query.ResponseChannel <- ReverseResponse{
				Err: err,
			}
			continue
		}

		results := searcher.Skip(query.Skip).Take(query.Take)

		query.ResponseChannel <- ReverseResponse{
			Results: results,
			Err:     nil,
		}
		searcher.Close()
	}

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

	needleIndex := ipconv.RoundDecIP(needle.Min, 10)

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

	scanner, file, err := getScanner(inputFileName, pos)
	if err != nil {
		return nil, err
	}

	reverseSearch := ReverseSearch{
		file:       file,
		needle:     needle,
		scanner:    scanner,
		query: query,
		foundFirst: false,
	}

	return &reverseSearch, nil

}

func (rs *ReverseSearch) Next() bool {
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case <-timeout:
			fmt.Printf("TIMEOUT ON %s\n", rs.query)
			rs.err = errors.New("timeout retrieving entry")
			return false
		default:
			break
		}

		if !rs.scanner.Scan() {
			rs.err = io.EOF
			return false
		}

		delimPos := bytes.IndexByte(rs.scanner.Bytes(), ',')
		if delimPos == -1 {
			continue
		}
		candidateIPv4, err := strconv.ParseUint(string(rs.scanner.Bytes()[:delimPos]), 10, 32)
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
		rs.reverseResult = reconstructReverseResult(candidateUInt32, string(rs.scanner.Bytes()[delimPos+1:]))
		return true
	}
}

func (rs *ReverseSearch) Skip(size int) *ReverseSearch {
	for i := 0; i < size; i++ {
		if !rs.Next() {
			break
		}
	}
	return rs
}

func (rs *ReverseSearch) Take(size int) map[string][]string {
	resultsMap := map[string][]string{}
	for i := 0; i < size; i++ {
		if !rs.Next() {
			break
		}
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
