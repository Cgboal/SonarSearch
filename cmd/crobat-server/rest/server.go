package rest

import (
	"net/http"

	"fmt"
	"github.com/Cgboal/DomainParser"
	"github.com/cgboal/sonarsearch/pkg/search"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"strconv"
)

var dp parser.Parser

var reverseQueries chan search.ReverseQuery
var domainQueries chan search.DomainQuery

func paginationHelper(c *gin.Context) (int, int) {
	limitString := c.Query("limit")
	pageString := c.Query("page")

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		limit = 100000
	}
	page, err := strconv.Atoi(pageString)
	if err != nil {
		page = 1
	}

	if page == 0 {
		page = 1
	}

	skip := (page - 1) * limit
	return skip, limit

}

func FindSubdomains(c *gin.Context) {
	query := c.Param("domain")
	skip, take := paginationHelper(c)
	responseChan := make(chan search.DomainResponse, 1)

	defer close(responseChan)

	domainQueries <- search.DomainQuery{Query: query, Take: take, Skip: skip, ResponseChannel: responseChan, NeedleFunc: search.FullDomainNeedle}

	response := <- responseChan

	if response.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": response.Err.Error()})
		return
	}
	c.JSON(http.StatusOK, response.Subdomains)
}

func FindAll(c *gin.Context) {
	searcher, err := search.NewDomainSearch(viper.GetString("domain_file"), c.Param("domain"), search.DomainNeedle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer searcher.Close()
	skip, limit := paginationHelper(c)
	subdomains := searcher.Skip(skip).Take(limit)
	c.JSON(http.StatusOK, subdomains)
}

func FindTLDs(c *gin.Context) {
	searcher, err := search.NewDomainSearch(viper.GetString("domain_file"), c.Param("domain"), search.DomainNeedle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer searcher.Close()
	skip, limit := paginationHelper(c)
	subdomains := searcher.Skip(skip).Take(limit)
	uniqueTLDs := map[string]struct{}{}
	for _, subdomain := range subdomains {
		domain := dp.ParseDomain(subdomain)
		fullDomain := fmt.Sprintf("%s.%s", domain.Domain, domain.TLD)
		_, exists := uniqueTLDs[fullDomain]
		if !exists {
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
	skip, take := paginationHelper(c)
	responseChan := make(chan search.ReverseResponse, 1)

	defer close(responseChan)

	reverseQueries <- search.ReverseQuery{Query: query, Take: take, Skip: skip, ResponseChannel: responseChan}

	response := <-responseChan

	if response.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": response.Err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Results[query])
}

func ReverseDNSCIDR(c *gin.Context) {
	query := fmt.Sprintf("%s/%s", c.Param("ip"), c.Param("cidr"))
	skip, take := paginationHelper(c)
	responseChan := make(chan search.ReverseResponse, 1)

	defer close(responseChan)

	reverseQueries <- search.ReverseQuery{Query: query, Take: take, Skip: skip, ResponseChannel: responseChan}

	response := <-responseChan

	if response.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": response.Err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Results)
	return
}

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	reverseQueries = make(chan search.ReverseQuery, 1)
	domainQueries = make(chan search.DomainQuery, 1)

	search.NewReversePool(reverseQueries)
	search.NewDomainPool(domainQueries)

	dp = parser.NewDomainParser()

	r.GET("/subdomains/:domain", FindSubdomains)
	r.GET("/tlds/:domain", FindTLDs)
	r.GET("/all/:domain", FindAll)
	r.GET("/reverse/:ip", ReverseDNS)
	r.GET("/reverse/:ip/:cidr", ReverseDNSCIDR)

	return r
}
