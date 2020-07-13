package main

import (
	"strings"
)

type SonarDomain struct {
	Name string `json:"name"`
	Subdomain string `json:"subdomain"`
	Domain string `json:"domain"`
	Tld string `json:"tld"`
	Value string `json:"value"`
	Type string `json:"type"`
}

func (d *SonarDomain) GetFullDomain() string {
	if d.Subdomain != "" {
		return strings.Join([]string{d.Subdomain, d.Domain, d.Tld}, ".")
	} else {
		return strings.Join([]string{d.Domain, d.Tld}, ".")
	}
}

func (d *SonarDomain) GetFQDN() string {
	return strings.Join([]string{d.Domain, d.Tld}, ".")
}
