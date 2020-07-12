package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	crobat "github.com/Cgboal/SonarSearch/proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"crypto/tls"
	"google.golang.org/grpc/credentials"
)

type SubdomainResponse []string
type AllResponse []map[string]string
type TldResponse []string

func (r SubdomainResponse) OutputPlain() {
	for _, domain := range r {
		fmt.Println(domain)
	}
}

func (r SubdomainResponse) OutputJSON() {
	json_out, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(json_out))
}

func (r AllResponse) OutputPlain() {
	for _, domain := range r {
		fmt.Println(domain["name"])
	}
}

func (r AllResponse) OutputJSON() {
	json_out, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(json_out))
}

func (r TldResponse) OutputPlain() {
	for _, domain := range r {
		fmt.Println(domain)
	}
}

func (r TldResponse) OutputJSON() {
	json_out, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(json_out))
}

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

func (c *CrobatClient) GetSubdomains(domain string, outputType string) {
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

func (c *CrobatClient) GetTlds(domain string, outputType string) {
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

func main() {
	domain_sub := flag.String("s", "", "Get subdomains for this value")
	domain_tld := flag.String("t", "", "Get tlds for this value")
	format := flag.String("f", "plain", "Set output format (json/plain)")

	flag.Parse()

	client := NewCrobatClient()
	if *domain_sub != "" {
		client.GetSubdomains(*domain_sub, *format)
	} else if *domain_tld != "" {
		client.GetTlds(*domain_tld, *format)
	}

}
