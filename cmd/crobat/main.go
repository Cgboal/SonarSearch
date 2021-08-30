package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	crobat "github.com/cgboal/sonarsearch/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type CrobatClient struct {
	conn   *grpc.ClientConn
	client crobat.CrobatClient
}

type UniqueStringSlice map[string]struct{}

func (slice *UniqueStringSlice) Append(value string) bool {
	if _, ok := (*slice)[value]; !ok {
		(*slice)[value] = struct{}{}
		return true
	}
	return false
}

func ProcessArg(arg string) (args []string) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	fileName := fmt.Sprintf("%s/%s", dir, arg)
	if _, err := os.Stat(fileName); err == nil {
		file, _ := os.Open(fileName)
		defer file.Close()
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			args = append(args, scanner.Text())
		}
	} else {
		args = strings.Split(arg, " ")
	}

	return args
}

func NewCrobatClient() CrobatClient {
	config := &tls.Config{}
	conn, err := grpc.Dial("crobat-rpc.omnisint.io:443", grpc.WithTransportCredentials(credentials.NewTLS(config)))
	if err != nil {
		log.Fatal(err)
	}

	client := crobat.NewCrobatClient(conn)
	return CrobatClient{
		conn:   conn,
		client: client,
	}
}

func (c *CrobatClient) GetSubdomains(arg string, resultsChan chan string) {
	args := ProcessArg(arg)
	defer close(resultsChan)
	for _, domain := range args {
		query := &crobat.QueryRequest{
			Query: domain,
		}

		stream, err := c.client.GetSubdomains(context.Background(), query)
		if err != nil {
			log.Fatal(err)
		}

		for {
			domain, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatal(err)
			}
			resultsChan <- domain.Domain
		}
	}

}

func (c *CrobatClient) GetTlds(arg string, resultsChan chan string) {
	args := ProcessArg(arg)
	defer close(resultsChan)
	for _, domain := range args {

		query := &crobat.QueryRequest{
			Query: domain,
		}

		stream, err := c.client.GetTLDs(context.Background(), query)
		if err != nil {
			log.Fatal(err)
		}

		for {
			domain, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatal(err)
			}
			resultsChan <- domain.Domain
		}
	}

}

func (c *CrobatClient) ReverseDNS(arg string, resultsChan chan string) {
	args := ProcessArg(arg)
	defer close(resultsChan)
	for _, ipv4 := range args {

		query := &crobat.QueryRequest{
			Query: ipv4,
		}

		stream, err := c.client.ReverseDNS(context.Background(), query)
		if err != nil {
			log.Fatal(err)
		}

		for {
			domain, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatal(err)
			}
			resultsChan <- domain.Domain
		}

	}
}

func (c *CrobatClient) ReverseDNSRange(arg string, resultsChan chan string) {
	args := ProcessArg(arg)
	defer close(resultsChan)
	for _, ipv4 := range args {
		query := &crobat.QueryRequest{
			Query: ipv4,
		}

		stream, err := c.client.ReverseDNSRange(context.Background(), query)
		if err != nil {
			log.Fatal(err)
		}

		for {
			domain, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if domain == nil {
				continue
			}
			resultsChan <- domain.Domain
		}

	}
}

func main() {
	domain_sub := flag.String("s", "", "Get subdomains for this value. Supports files and quoted lists")
	domain_tld := flag.String("t", "", "Get tlds for this value. Supports files and quoted lists")
	reverse_dns := flag.String("r", "", "Perform reverse lookup on IP address or CIDR range. Supports files and quoted lists")
	unique_sort := flag.Bool("u", false, "Ensures results are unique, may cause instability on large queries due to RAM requirements")

	resultsChan := make(chan string)
	var wg sync.WaitGroup

	flag.Parse()

	wg.Add(1)
	go func() {
		defer wg.Done()
		uniqueSlice := UniqueStringSlice{}
		for result := range resultsChan {
			if *unique_sort {
				if uniqueSlice.Append(result) {
					fmt.Println(result)
				}
			} else {
				fmt.Println(result)
			}
		}
	}()

	client := NewCrobatClient()
	if *domain_sub != "" {
		client.GetSubdomains(*domain_sub, resultsChan)
	} else if *domain_tld != "" {
		client.GetTlds(*domain_tld, resultsChan)
	} else if *reverse_dns != "" {
		if !strings.Contains(*reverse_dns, "/") {
			client.ReverseDNS(*reverse_dns, resultsChan)
		} else {
			client.ReverseDNSRange(*reverse_dns, resultsChan)
		}
	}

	wg.Wait()

}
