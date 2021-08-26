## SonarSearch
<p align="center">
  <img width="30%" height="30%" src="https://sonar.omnisint.io/img/crobat.png">
</p>
This repository contains all the code needed to create index Rapid7's Project Sonar Forward DNS lookup datasets into a MongoDB database, and query them in a time efficient fashion. 

An instance of this API (Crobat) is online at the following URL: 

> https://sonar.omnisint.io

### Crobat
Crobat is a command line utility designed to allow easy querying of the Crobat API. To install the client, run the following command: 
``` normal
$ go get github.com/Cgboal/SonarSearch/crobat
```

A Docker container is also available: 
```normal
$ docker run cgboal/crobat:latest -h 
```

By default, Crobat will return a list of result in plain-text, however, JSON output is also supported. 

Below is a full list of command line flags:
``` normal
$ crobat -h                                                                                                                                                                      
Usage of crobat:
  -r string
    	Perform reverse lookup on IP address or CIDR range. Supports files and quoted lists
  -s string
    	Get subdomains for this value. Supports files and quoted lists
  -t string
    	Get tlds for this value. Supports files and quoted lists
  -u	Ensures results are unique, may cause instability on large queries due to RAM requirements
```

Additionally, it is now possible to pass either file names, or quoted lists ('example.com example.co.uk') as the value for each flag in order to specify multiple domains/ranges.

### Crobat API

Currently, Project Crobat offers two APIs. The first of these is a REST API, with the following endpoints: 

``` normal
/subdomains/{domain} - All subdomains for a given domain
/tlds/{domain} - All tlds found for a given domain
/all/{domain} - All results across all tlds for a given domain
/reverse/{ip} - Reverse DNS lookup on IP address
/reverse/{ip}/{mask} - Reverse DNS lookup of a CIDR range
```

Additionally, Project Crobat offers a gRPC API which is used by the client to stream results over HTTP/2. Thus, it is recommended that the client is used for large queries as it reduces both query execution times, and server load. Also, unlike the REST API, there is no limit to the size of specified when performing reverse DNS lookups. 

No authentication is required to use the API, nor special headers, so go nuts. 

However, the REST API does have pagination. Currently pages are limited to 10k results per page. To request pages, add `?page=X` to the request, where `X` is the page number. The gRPC API does not require pagination as results are streamed to the client. 


### Third-Party SDKs

* [Crystal SDK and CLI tool complete with Docker images](https://github.com/PercussiveElbow/crobat-sdk-crystal) made by [@mil0sec](https://twitter.com/mil0sec)

### Contributing 
If you wish to contribute a SDK written in other languages, shoot me a DM on Twitter (@CalumBoal), or open an issue on this repository and I will provide a link to your repository in the Third-Party SDK's section of this readme. 
