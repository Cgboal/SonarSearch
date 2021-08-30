package search

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	parser "github.com/Cgboal/DomainParser"
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
	reader     *bufio.Reader
	subdomain  string
	err        error
	foundFirst bool
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

	pos := domainIndex[queryDomain.Domain]
	if pos == 0 {
		return nil, errors.New("no results found")
	}

	reader, file, err := getReader(inputFileName, pos)
	if err != nil {
		return nil, err
	}

	domainSearch := DomainSearch{
		file:      file,
		needle:    needle,
		needleLen: len(needle),
		reader:    reader,
	}

	return &domainSearch, nil

}

func (ds *DomainSearch) Next() bool {
	for {
		line, err := ds.reader.ReadBytes('\n')
		if err == io.EOF {
			ds.err = err
		}

		if string(line[:ds.needleLen]) != ds.needle {
			if ds.foundFirst {
				return false
			} else {
				continue
			}
		}

		ds.foundFirst = true
		ds.subdomain = reconstructDomainLine(line)
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

func (ds *DomainSearch) Collect() []string {
	subdomains := []string{}
	for ds.Next() {
		subdomains = append(subdomains, ds.Text())

		if ds.err == io.EOF {
			break
		}
	}

	return subdomains
}

func getReader(fileName string, pos int64) (*bufio.Reader, *os.File, error) {
	inputFile, err := os.Open(fileName)
	if err != nil {
		return nil, inputFile, err
	}

	inputFile.Seek(int64(pos), 0)

	reader := bufio.NewReader(inputFile)
	return reader, inputFile, nil
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
	if parts[2] == "\n" {
		subdomain = fmt.Sprintf("%s.%s", parts[0], parts[1])
	} else {
		parts[2] = strings.TrimRight(parts[2], "\n")
		subdomain = fmt.Sprintf("%s.%s.%s", parts[2], parts[0], parts[1])
	}

	return subdomain

}
