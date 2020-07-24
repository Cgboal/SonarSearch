package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	crobat "github.com/Cgboal/SonarSearch/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"strings"
)

type CrobatClient struct {
	conn   *grpc.ClientConn
	client crobat.CrobatClient
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

func (c *CrobatClient) GetSubdomains(domain string) {
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
		fmt.Println(domain.Domain)
	}

}

func (c *CrobatClient) GetTlds(domain string) {
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
		fmt.Println(domain.Domain)
	}

}

func (c *CrobatClient) ReverseDNS(ipv4 string) {
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
		fmt.Println(domain.Domain)
	}

}

func (c *CrobatClient) ReverseDNSRange(ipv4 string) {
	query := &crobat.QueryRequest{
		Query: ipv4,
	}

	stream, err := c.client.ReverseDNSRange(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	for {
		result, err := stream.Recv()
		if err == io.EOF {
			break
		}
		jsonResults, _ := json.MarshalIndent(*result, "", "    ")
		fmt.Printf("%s", jsonResults)
	}

}

func main() {
	domain_sub := flag.String("s", "", "Get subdomains for this value")
	domain_tld := flag.String("t", "", "Get tlds for this value")
	reverse_dns := flag.String("r", "", "Perform reverse lookup on IP address or CIDR range")

	flag.Parse()

	client := NewCrobatClient()
	if *domain_sub != "" {
		client.GetSubdomains(*domain_sub)
	} else if *domain_tld != "" {
		client.GetTlds(*domain_tld)
	} else if *reverse_dns != "" {
		if !strings.Contains(*reverse_dns, "/") {
			client.ReverseDNS(*reverse_dns)
		} else {
			client.ReverseDNSRange(*reverse_dns)
		}
	}

}
