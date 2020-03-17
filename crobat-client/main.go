package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"time"
	"strings"
	"github.com/sethvargo/go-diceware/diceware"
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
	config := load_config()
	url := fmt.Sprintf("https://%s:%s", config["host"], config["port"])
	return CrobatClient{
		http_client: client,
		url:         url,
	}
}

func (c *CrobatClient) GetPage(uri string) func() *http.Response {
	page := 0
	return func() *http.Response {
		config := load_config()
		req, err := http.NewRequest("GET", fmt.Sprintf("%s%s?page=%d", c.url, uri, page), nil)
		client := &http.Client{Timeout: time.Second * 10}

		req.Header.Set("User-Agent", "Crobat: 1.1")
		req.Header.Set("X-id", config["uid"])

		resp, err := client.Do(req)
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

func generate_uid() string {
	list, _ := diceware.Generate(4)
	var capitalized strings.Builder
	for i := 0; i < len(list); i++ {
		if i != len(list) - 1 {
			capitalized.WriteString(strings.Title(list[i] + "-"))
		} else {
			capitalized.WriteString(strings.Title(list[i]))
		}
		
	}
 	return capitalized.String()
}

func load_config() map[string]string {
	usr, err := user.Current()
	path := fmt.Sprintf("%s/.crobatrc", usr.HomeDir)
	jsonFile, err := os.Open(path)

	if err != nil {
		fmt.Println("Unable to load connection details from ~/.crobatrc, did you use --init?")
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	defer jsonFile.Close()
	var result map[string]string
	json.NewDecoder(jsonFile).Decode(&result)
	return result
}

func main() {
	initialize := flag.Bool("init", false, "Initialize config and auth file")
	domain_sub := flag.String("s", "", "Get subdomains for this value")
	domain_tld := flag.String("t", "", "Get tlds for this value")
	domain_all := flag.String("all", "", "Get all data for this query")
	format := flag.String("f", "plain", "Set output format (json/plain)")

	flag.Parse()

	if *initialize {
		var host string
		var port string
		usr, _ := user.Current()
		path := fmt.Sprintf("%s/.crobatrc", usr.HomeDir)
		config := make(map[string]string)
		fmt.Println("Initializing ~/.crobatrc")
		fmt.Println("Warnining: this will overwrite existing data in ~/.crobatrc, use ctrl+c to abort.")
		fmt.Printf("Host: ")
		fmt.Scan(&host)
		fmt.Printf("Port: ")
		fmt.Scan(&port)

		config["host"] = host
		config["port"] = port
		config["uid"] = generate_uid()

		str, _ := json.MarshalIndent(config, "", "  ")

		err := ioutil.WriteFile(path, str, 0644)
		if err == nil {
			fmt.Println("Saved to ~/.crobatrc successfully")
		} else {
			fmt.Println("Saving ~/.crobatrc failed")
			fmt.Println("Error:", err)
		}
	}

	client := NewCrobatClient()
	if *domain_sub != "" {
		client.GetSubdomains(*domain_sub, *format)
	} else if *domain_all != "" {
		client.GetAll(*domain_all, *format)
	} else if *domain_tld != "" {
		client.GetTlds(*domain_tld, *format)
	}

}
