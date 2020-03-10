package main

func (s *server) routes() {
	s.Router.HandleFunc("/subdomains/{domain}", s.LookupHandler("subdomains"))
	s.Router.HandleFunc("/all/{domain}", s.LookupHandler("all"))
	s.Router.HandleFunc("/tlds/{domain}", s.LookupHandler("tlds"))
}
