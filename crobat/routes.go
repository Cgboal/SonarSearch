package main

func (s *server) routes() {
	s.Router.HandleFunc("/subdomains/{domain}", s.SubdomainHandler())
	s.Router.HandleFunc("/all/{domain}", s.AllHandler())
	s.Router.HandleFunc("/tlds/{domain}", s.TldHandler())
}
