package main

import (
	"context"
	"encoding/json"
	"github.com/Cgboal/DomainParser"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"strings"
	"time"
)

type server struct {
	db     *mongo.Client
	Router *mux.Router
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
	}
	server.routes()

	return server
}

func internal_error(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"Message": "` + err.Error() + `"}`))
}

func json_response(w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(v)
}

func (s *server) LookupHandler(a string) http.HandlerFunc {
	collection := s.db.Database("sonar").Collection("A")
	dp := parser.NewDomainParser()
	action := a

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		vars := mux.Vars(r)
		domain_index := strings.Split(vars["domain"], ".")[0]
		query := bson.M{"domain_index": domain_index}
		cur, err := collection.Find(ctx, query)
		if err != nil {
			internal_error(w, err)
			return
		}

		defer cur.Close(ctx)

		switch action {
		case "subdomains":
			var domains []string
			for cur.Next(ctx) {
				var domain SonarDomain
				cur.Decode(&domain)
				if dp.GetFQDN(domain.Name) == vars["domain"] {
					domains = append(domains, domain.Name)
				}
			}
			json_response(w, domains)
		case "all":
			var domains []SonarDomain
			for cur.Next(ctx) {
				var domain SonarDomain
				cur.Decode(&domain)
				domains = append(domains, domain)
			}
			json_response(w, domains)
		case "tlds":
			domains_map := make(map[string]struct{})
			for cur.Next(ctx) {
				var domain SonarDomain
				cur.Decode(&domain)
				fqdn := dp.GetFQDN(domain.Name)
				if _, ok := domains_map[fqdn]; !ok {
					domains_map[fqdn] = struct{}{}
				}
			}
			domains := make([]string, len(domains_map))
			i := 0
			for k := range domains_map {
				domains[i] = k
				i++
			}
			json_response(w, domains)
		}

		if err := cur.Err(); err != nil {
			internal_error(w, err)
			return
		}

	}
}
