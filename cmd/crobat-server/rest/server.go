package rest

import (
	"net/http"

	"github.com/cgboal/sonarsearch/pkg/search"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"fmt"
	"github.com/Cgboal/DomainParser"
)

var dp parser.Parser

func FindSubdomains(c *gin.Context) {
	searcher, err := search.NewDomainSearch(viper.GetString("domain_file"), c.Param("domain"), search.FullDomainNeedle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer searcher.Close()
	subdomains := searcher.Collect()
	c.JSON(http.StatusOK, subdomains)
}

func FindAll(c *gin.Context) {
	searcher, err := search.NewDomainSearch(viper.GetString("domain_file"), c.Param("domain"), search.DomainNeedle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer searcher.Close()
	subdomains := searcher.Collect()
	c.JSON(http.StatusOK, subdomains)
}

func FindTLDs(c *gin.Context) {
	searcher, err := search.NewDomainSearch(viper.GetString("domain_file"), c.Param("domain"), search.DomainNeedle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer searcher.Close()
	subdomains := searcher.Collect()
	uniqueTLDs := map[string]struct{}{}
	for _, subdomain := range subdomains {
		domain := dp.ParseDomain(subdomain)
		fullDomain := fmt.Sprintf("%s.%s", domain.Domain, domain.TLD)
		_, exists := uniqueTLDs[fullDomain]; if !exists {
			uniqueTLDs[fullDomain] = struct{}{}
		}
	}

	results := []string{}
	for domain := range uniqueTLDs {
		results = append(results, domain)
	}

	c.JSON(http.StatusOK, results)

}

func ReverseDNS(c *gin.Context) {
	query := c.Param("ip")
	searcher, err := search.NewReverseSearch(viper.GetString("reverse_file"), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer searcher.Close()
	results := searcher.Collect()
	c.JSON(http.StatusOK, results[query])
	return
}

func ReverseDNSCIDR(c *gin.Context) {
	query := fmt.Sprintf("%s/%s", c.Param("ip"), c.Param("cidr"))
	searcher, err := search.NewReverseSearch(viper.GetString("reverse_file"), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer searcher.Close()
	results := searcher.Collect()
	c.JSON(http.StatusOK, results)
	return
}

func NewServer() *gin.Engine {
	r := gin.Default()

	dp = parser.NewDomainParser()

	r.GET("/subdomains/:domain", FindSubdomains)
	r.GET("/tlds/:domain", FindTLDs)
	r.GET("/all/:domain", FindAll)
	r.GET("/reverse/:ip", ReverseDNS)
	r.GET("/reverse/:ip/:cidr", ReverseDNSCIDR)

	return r
}
