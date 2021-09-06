package search

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"time"

	parser "github.com/Cgboal/DomainParser"
	"github.com/spf13/viper"
)

var dp parser.Parser

func init() {
	dp = parser.NewDomainParser()
}

type domainNeedleFunc func(parser.Domain) (string, error)

type DomainSearch struct {
	file       *os.File
	needle     string
	needleLen  int
	scanner    *bufio.Scanner
	subdomain  string
	query      []byte
	err        error
	foundFirst bool
}

type DomainResponse struct {
	Subdomains []string
	Err        error
}

type DomainQuery struct {
	Query           string
	NeedleFunc      domainNeedleFunc
	Take            int
	Skip            int
	ResponseChannel chan DomainResponse
}

func NewDomainPool(requests <-chan DomainQuery) {
	for i := 0; i < 5; i++ {
		go startDomainWorker(requests)
	}
}

func startDomainWorker(requests <-chan DomainQuery) {
	for query := range requests {
		searcher, err := NewDomainSearch(viper.GetString("domain_file"), query.Query, query.NeedleFunc)
		if err != nil {
			query.ResponseChannel <- DomainResponse{
				Err: err,
			}
			continue
		}

		subdomains := searcher.Skip(query.Skip).Take(query.Take)

		query.ResponseChannel <- DomainResponse{
			Subdomains: subdomains,
			Err:        nil,
		}
		searcher.Close()
	}

}

func NewDomainSearch(inputFileName string, query string, needleFunc domainNeedleFunc) (*DomainSearch, error) {
	if query == "" {
		return nil, errors.New("query cannot be blank")
	}

	queryDomain := dp.ParseDomain(query)

	needle, err := needleFunc(queryDomain)
	if err != nil {
		return nil, err
	}

	pos, err := getPos(queryDomain.Domain)

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

	domainSearch := DomainSearch{
		file:      file,
		needle:    needle,
		needleLen: len(needle),
		query:     []byte(query),
		scanner:   scanner,
	}

	return &domainSearch, nil

}

func (ds *DomainSearch) Next() bool {
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case <-timeout:
			fmt.Printf("TIMEOUT ON %s\n", ds.query)
			ds.err = errors.New("timeout retrieving entry")
			return false
		default:
			break
		}

		if !ds.scanner.Scan() {
			ds.err = io.EOF
			return false
		}

		if len(ds.scanner.Bytes()) < ds.needleLen {
			continue
		}

		if string(ds.scanner.Bytes()[:ds.needleLen]) != ds.needle {
			if ds.foundFirst {
				return false
			} else {
				continue
			}
		}

		ds.foundFirst = true
		ds.subdomain = reconstructDomainLine(ds.scanner.Bytes())
		return true
	}
}

func (ds *DomainSearch) Text() string {
	return ds.subdomain
}

func (ds *DomainSearch) Close() {
	ds.file.Close()
}

func (ds *DomainSearch) Error() error {
	return ds.err
}

func (ds *DomainSearch) Skip(size int) *DomainSearch {
	for i := 0; i < size; i++ {
		if !ds.Next() {
			break
		}
	}

	return ds
}

func (ds *DomainSearch) Take(size int) []string {
	subdomains := []string{}
	for i := 0; i < size; i++ {
		if !ds.Next() {
			break
		}

		subdomains = append(subdomains, ds.Text())

		if ds.err == io.EOF {
			break
		}
	}

	return subdomains
}

func getScanner(fileName string, pos int64) (*bufio.Scanner, *os.File, error) {
	inputFile, err := os.Open(fileName)
	if err != nil {
		return nil, inputFile, err
	}

	inputFile.Seek(int64(pos), 0)

	scanner := bufio.NewScanner(inputFile)
	scanner.Buffer(make([]byte, 10240), 10240)
	return scanner, inputFile, nil
}

func FullDomainNeedle(queryDomain parser.Domain) (string, error) {
	needle := fmt.Sprintf("%s,%s,", queryDomain.Domain, queryDomain.TLD)
	return needle, nil

}

func DomainNeedle(queryDomain parser.Domain) (string, error) {
	needle := fmt.Sprintf("%s,", queryDomain.Domain)
	return needle, nil
}

func reconstructDomainLine(line []byte) string {
	lineStr := string(line)

	parts := strings.Split(lineStr, ",")

	var subdomain string
	if len(parts[2]) == 0 {
		subdomain = fmt.Sprintf("%s.%s", parts[0], parts[1])
	} else {
		parts[2] = strings.TrimRight(parts[2], "\n")
		subdomain = fmt.Sprintf("%s.%s.%s", parts[2], parts[0], parts[1])
	}

	return subdomain

}
