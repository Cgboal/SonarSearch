package main

import (
	"context"
	"fmt"
	parser "github.com/Cgboal/DomainParser"
	crobat "github.com/Cgboal/SonarSearch/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"time"
)

type crobatServer struct {
	db *mongo.Client
	dp parser.Parser
}

func NewRPCServer() crobatServer {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	server := crobatServer{
		db: client,
		dp: parser.NewDomainParser(),
	}

	return server
}

func (s *crobatServer) GetSubdomains(query *crobat.QueryRequest, stream crobat.Crobat_GetSubdomainsServer) error {
	collection := s.db.Database("sonar").Collection("A")
	queryCtx, _ := context.WithTimeout(context.Background(), 120*time.Second)
	domain := s.dp.GetDomain(query.Query)
	tld := s.dp.GetTld(query.Query)
	mongoQuery := bson.M{"domain": domain, "tld": tld}
	opts := options.Find().SetProjection(bson.D{{"subdomain", 1}, {"domain", 1}, {"tld", 1}})
	cur, err := collection.Find(queryCtx, mongoQuery, opts)
	if err != nil {
		return err
	}
	defer cur.Close(queryCtx)
	for cur.Next(queryCtx) {
		var domain SonarDomain
		cur.Decode(&domain)
		reply := &crobat.Domain{
			Domain: domain.GetFullDomain(),
			Ipv4:   domain.Value,
		}
		if err := stream.Send(reply); err != nil {
			return err
		}
	}

	return nil
}

func (s *crobatServer) GetTLDs(query *crobat.QueryRequest, stream crobat.Crobat_GetTLDsServer) error {
	collection := s.db.Database("sonar").Collection("A")
	ctx, _ := context.WithTimeout(context.Background(), 120*time.Second)
	domain := query.Query
	mongoQuery := bson.M{"domain": domain}
	values, err := collection.Distinct(ctx, "tld", mongoQuery)
	if err != nil {
		return err
	}

	for _, tld := range values {
		reply := &crobat.Domain{
			Domain: fmt.Sprintf("%s.%s", query.Query, tld),
		}
		if err := stream.Send(reply); err != nil {
			return err
		}
	}
	return nil

}

func (s *crobatServer) ReverseDNS(query *crobat.QueryRequest, stream crobat.Crobat_ReverseDNSServer) error {
	collection := s.db.Database("sonar").Collection("A")
	ctx, _ := context.WithTimeout(context.Background(), 120*time.Second)
	mongoQuery := bson.M{"value": query.Query}
	cur, err := collection.Find(ctx, mongoQuery)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var domain SonarDomain
		cur.Decode(&domain)
		reply := &crobat.Domain{
			Domain: domain.GetFullDomain(),
		}
		if err := stream.Send(reply); err != nil {
			return err
		}
	}
	return nil
}

func (s *crobatServer) ReverseDNSRange(ctx context.Context, query *crobat.QueryRequest) (*crobat.ReverseReply, error) {
	collection := s.db.Database("sonar").Collection("A")
	mongoCtx, _ := context.WithTimeout(context.Background(), 120*time.Second)
	cidr := query.Query
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return &crobat.ReverseReply{}, err
	}
	reply := &crobat.ReverseReply{}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		reverse_results, err := func(ip net.IP) ([]*crobat.Domain, error) {
			mongoQuery := bson.M{"value": ip.String()}
			cur, err := collection.Find(mongoCtx, mongoQuery)
			if err != nil {
				return nil, err
			}
			defer cur.Close(ctx)
			var results []*crobat.Domain
			for cur.Next(ctx) {
				var domain SonarDomain
				cur.Decode(&domain)
				result := &crobat.Domain{
					Domain: domain.GetFullDomain(),
					Ipv4:   domain.Value,
				}
				results = append(results, result)
			}
			return results, nil
		}(ip)
		if err != nil {
			return &crobat.ReverseReply{}, err
		}
		if reverse_results != nil {
			result := &crobat.ReverseResult{
				Ip:      ip.String(),
				Domains: reverse_results,
			}
			reply.Results = append(reply.Results, result)
		}
	}
	return reply, nil
}
