package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
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
	http_client http.Client
	url         string
}

func NewCrobatClient() CrobatClient {
	client := http.Client{Timeout: 60 * time.Second}
	url := "https://sonar.omnisint.io/"
	return CrobatClient{
		http_client: client,
		url:         url,
	}
}

func (c *CrobatClient) GetPage(uri string) func() *http.Response {
	page := 0
	return func() *http.Response {
		resp, err := c.http_client.Get(fmt.Sprintf("%s%s?page=%d", c.url, uri, page))
		if err != nil {
			log.Fatal(err)
		}
		page++
		return resp
	}
}

func (c *CrobatClient) GetSubdomains(domain string, outputType string) {
	var subdomains SubdomainResponse
	var results SubdomainResponse

	getPage := c.GetPage("/subdomains/" + domain)
	for {
		resp := getPage()
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&subdomains)
		if subdomains == nil {
			break
		}
		if outputType != "json" {
			subdomains.OutputPlain()
		} else {
			results = append(results, subdomains...)
		}
	}

	if outputType != "json" {
		return
	}
	results.OutputJSON()
}

func (c *CrobatClient) GetAll(domain string, outputType string) {
	var tlds AllResponse
	var results AllResponse
	getPage := c.GetPage("/all/" + domain)
	for {
		resp := getPage()
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&tlds)
		if tlds == nil {
			break
		}
		if outputType != "json" {
			tlds.OutputPlain()
		} else {
			results = append(results, tlds...)
		}
	}

	if outputType != "json" {
		return
	}
	results.OutputJSON()
}

func (c *CrobatClient) GetTlds(domain string, outputType string) {
	var subdomains TldResponse
	var results TldResponse
	getPage := c.GetPage("/tlds/" + domain)
	resp := getPage()
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&subdomains)

	if outputType == "json" {
		results.OutputJSON()
		return
	} else if outputType == "plain" {
		subdomains.OutputPlain()
	}
}

func main() {
	domain_sub := flag.String("s", "", "Get subdomains for this value")
	domain_tld := flag.String("t", "", "Get tlds for this value")
	domain_all := flag.String("all", "", "Get all data for this query")
	format := flag.String("f", "plain", "Set output format (json/plain)")

	flag.Parse()

	client := NewCrobatClient()
	if *domain_sub != "" {
		client.GetSubdomains(*domain_sub, *format)
	} else if *domain_all != "" {
		client.GetAll(*domain_all, *format)
	} else if *domain_tld != "" {
		client.GetTlds(*domain_tld, *format)
	}

}
