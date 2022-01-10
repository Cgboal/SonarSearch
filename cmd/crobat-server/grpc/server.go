package grpc

import (
	"fmt"
	parser "github.com/Cgboal/DomainParser"
	"github.com/cgboal/sonarsearch/pkg/search"
	crobat "github.com/cgboal/sonarsearch/proto"
	"github.com/spf13/viper"
	"io"
)

var dp parser.Parser

type CrobatServer struct{
	crobat.UnimplementedCrobatServer
}

func (s *CrobatServer) GetSubdomains(query *crobat.QueryRequest, stream crobat.Crobat_GetSubdomainsServer) error {
	searcher, err := search.NewDomainSearch(viper.GetString("domain_file"), query.Query, search.FullDomainNeedle)
	if err != nil {
		return err
	}
	defer searcher.Close()
	for searcher.Next() {
		domain := searcher.Text()
		reply := &crobat.Domain{
			Domain: domain,
		}
		if err := stream.Send(reply); err != nil {
			return err
		}

		if searcher.Error() == io.EOF {
			break
		}
	}

	return nil
}

func (s *CrobatServer) GetTLDs(query *crobat.QueryRequest, stream crobat.Crobat_GetTLDsServer) error {
	searcher, err := search.NewDomainSearch(viper.GetString("domain_file"), query.Query, search.DomainNeedle)
	if err != nil {
		return err
	}
	defer searcher.Close()
	uniqueTLDs := map[string]struct{}{}
	for searcher.Next() {
		subdomain := searcher.Text()
		domain := dp.ParseDomain(subdomain)
		fullDomain := fmt.Sprintf("%s.%s", domain.Domain, domain.TLD)
		_, exists := uniqueTLDs[fullDomain]
		if !exists {
			uniqueTLDs[fullDomain] = struct{}{}
			reply := &crobat.Domain{
				Domain: fullDomain,
			}
			if err := stream.Send(reply); err != nil {
				return err
			}
		}
	}
	return nil

}

func (s *CrobatServer) ReverseDNS(query *crobat.QueryRequest, stream crobat.Crobat_ReverseDNSServer) error {
	searcher, err := search.NewReverseSearch(viper.GetString("reverse_file"), query.Query)
	if err != nil {
		return err
	}
	defer searcher.Close()
	for searcher.Next() {
		result := searcher.Result()
		reply := &crobat.Domain{
			Domain: result.Domain,
                        IPv4: result.IPv4,
		}
		if err := stream.Send(reply); err != nil {
			return err
		}

		if searcher.Error() == io.EOF {
			break
		}
	}
	return nil
}

func (s *CrobatServer) ReverseDNSRange(query *crobat.QueryRequest, stream crobat.Crobat_ReverseDNSRangeServer) error {
	searcher, err := search.NewReverseSearch(viper.GetString("reverse_file"), query.Query)
	if err != nil {
		return err
	}
	defer searcher.Close()
	for searcher.Next() {
		result := searcher.Result()
		reply := &crobat.Domain{
			Domain: result.Domain,
                        IPv4: result.IPv4,
		}
		if err := stream.Send(reply); err != nil {
			return err
		}

		if searcher.Error() == io.EOF {
			break
		}
	}
	return nil
}
