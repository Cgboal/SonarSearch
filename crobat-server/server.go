package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Cgboal/DomainParser"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type domainHandler = func(w http.ResponseWriter, r *http.Request, cur *mongo.Cursor, ctx context.Context)

type server struct {
	db     *mongo.Client
	Router *mux.Router
	dp     parser.Parser
}

func NewServer() server {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	server := server{
		db:     client,
		Router: mux.NewRouter(),
		dp:     parser.NewDomainParser(),
	}
	server.routes()

	return server
}

func internal_error(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"Message": "` + err.Error() + `"}`))
}

func unauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"Message": "Authentication Required"`))
}

func json_response(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(v)
}

func get_page_id(r *http.Request) (int, error) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "0"
	}

	page_int, err := strconv.Atoi(page)
	return page_int, err

}

func get_limit(r *http.Request) (int, error) {
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "10000"
	}

	limit_int, err := strconv.Atoi(limit)
	if limit_int > 10000 {
		limit_int = 10000
	}

	return limit_int, err

}

func (s *server) paginateDomains(ctx context.Context, page int, limit int, query bson.M, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	collection := s.db.Database("sonar").Collection("A")

	opts = append(opts, options.Find().SetLimit(int64(limit)))
	opts = append(opts, options.Find().SetSkip(int64(limit*page)))
	cur, err := collection.Find(ctx, query, opts...)

	if err != nil {
		return nil, err
	}

	return cur, nil

}

func (s *server) SubdomainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := context.WithTimeout(context.Background(), 120*time.Second)
		vars := mux.Vars(r)
		fullDomain := vars["domain"]
		domain := s.dp.GetDomain(fullDomain)
		tld := s.dp.GetTld(vars["domain"])
		query := bson.M{"domain": domain, "tld": tld}

		fullDomainParts := strings.Split(fullDomain, ".")
		filterResults := (len(fullDomainParts) >= 3)

		page, err := get_page_id(r)
		if err != nil {
			internal_error(w, err)
			return
		}
		limit, err := get_limit(r)
		if err != nil {
			internal_error(w, err)
			return
		}

		opts := options.Find().SetProjection(bson.D{{"subdomain", 1}, {"domain", 1}, {"tld", 1}})
		cur, err := s.paginateDomains(ctx, page, limit, query, opts)
		if err != nil {
			internal_error(w, err)
			return
		}
		defer cur.Close(ctx)
		var domains []string
		for cur.Next(ctx) {
			var domain SonarDomain
			cur.Decode(&domain)
			result := domain.GetFullDomain()
			if filterResults {
				if !strings.Contains(result, "."+fullDomain) {
					continue
				}
			}
			domains = append(domains, result)
		}
		json_response(w, domains)
	}

}

func (s *server) ReverseRangeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection := s.db.Database("sonar").Collection("A")
		ctx, _ := context.WithTimeout(context.Background(), 120*time.Second)
		vars := mux.Vars(r)

		maskSize, err := strconv.Atoi(vars["mask"])
		if err != nil {
			internal_error(w, err)
			return
		}
		fmt.Println(maskSize)
		if maskSize < 16 {
			internal_error(w, errors.New("If you want to request networks larger than a /16, pease use the command line client which streams the results thus reducing server load."))
			return
		}

		cidr := fmt.Sprintf("%s/%s", vars["ip"], vars["mask"])
		ip, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			internal_error(w, err)
			return
		}
		ips := bson.A{}
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
			ips = append(ips, ip.String())
		}
		query := bson.M{"value": bson.M{"$in": ips}}
		cur, err := collection.Find(ctx, query)
		if err != nil {
			internal_error(w, err)
			return
		}
		defer cur.Close(ctx)
		results := map[string][]string{}
		for cur.Next(ctx) {
			var domain SonarDomain
			cur.Decode(&domain)
			results[domain.Value] = append(results[domain.Value], domain.GetFullDomain())
		}
		json_response(w, results)
	}
}

func (s *server) ReverseHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection := s.db.Database("sonar").Collection("A")
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		vars := mux.Vars(r)
		query := bson.M{"value": vars["ip"]}
		cur, err := collection.Find(ctx, query)
		if err != nil {
			internal_error(w, err)
			return
		}
		defer cur.Close(ctx)
		var domains []string
		for cur.Next(ctx) {
			var domain SonarDomain
			cur.Decode(&domain)
			domains = append(domains, domain.GetFullDomain())
		}
		json_response(w, domains)
	}
}

func (s *server) TldHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		collection := s.db.Database("sonar").Collection("A")
		vars := mux.Vars(r)
		domain := s.dp.GetDomain(vars["domain"])
		query := bson.M{"domain": domain}

		values, err := collection.Distinct(ctx, "tld", query)
		if err != nil {
			internal_error(w, err)
			return
		}
		var domains []string
		for _, tld := range values {
			domains = append(domains, fmt.Sprintf("%s.%s", domain, tld))
		}
		json_response(w, domains)

	}
}

func (s *server) AllHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		vars := mux.Vars(r)
		domain := s.dp.GetDomain(vars["domain"])
		query := bson.M{"domain": domain}

		page, err := get_page_id(r)
		if err != nil {
			internal_error(w, err)
			return
		}

		limit, err := get_limit(r)
		if err != nil {
			internal_error(w, err)
			return
		}

		cur, err := s.paginateDomains(ctx, page, limit, query)
		defer cur.Close(ctx)
		var domains []SonarDomain
		for cur.Next(ctx) {
			var domain SonarDomain
			cur.Decode(&domain)
			domain.Name = domain.GetFullDomain()
			domains = append(domains, domain)
		}
		json_response(w, domains)

	}
}
