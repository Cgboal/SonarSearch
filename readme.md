## SonarSearch v2
<p align="center">
  <img width="30%" height="30%" src="https://sonar.omnisint.io/img/crobat.png">
</p>

### Attention!
Over a year ago Rapid7 revoked public access to their datasets, and thus the data hosted on the omnisint API became extremely out of date. In addition, due to the licensing changes around the data, our wonderful sponsor ZeroGuard was no longer able to support the project. As a result, it has been taken offline. However, I have released full instruction for running your own instance of the API, providing you can obtain a dataset. The instructions can be found at the bottom of the README.

-----

This repo contains all the tools needed to create a blazing fast API for Rapid7's Project Sonar dataset. It employs a custom indexing method in order to achieve fast lookups of both subdomains for a given domain, and domains which resolve to a given IP address. 

-----


An instance of this API (Crobat) is online at the following URL: 

> https://sonar.omnisint.io

### Crobat
Crobat is a command line utility designed to allow easy querying of the Crobat API. To install the client, run the following command: 
``` normal
$ go get github.com/cgboal/sonarsearch/cmd/crobat
```

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

### Third-Party SDKs

* [Crystal SDK and CLI tool complete with Docker images](https://github.com/PercussiveElbow/crobat-sdk-crystal) made by [@mil0sec](https://twitter.com/mil0sec)

### Contributing 
If you wish to contribute a SDK written in other languages, shoot me a DM on Twitter (@CalumBoal), or open an issue on this repository and I will provide a link to your repository in the Third-Party SDK's section of this readme. 


# SonarSearch Setup Instructions 

Setting up an instance of SonarSearch is reasonably straightforward. You will require a host to run the server on, this can be a VPS, or your own personal device. Regardless of the hosting option you choose, you will require 150-200GB of diskspace in order to store the datasets and indexes. 

There are two options for hosting the indexes (redis, or postgres). Redis requires ~20GB of RAM to hold the index, but it is quick to load the index, as well as query it. Postgres on the other hand does not use ram to hold the index, and thus has a much lower memory footprint. However, it will take longer to load the data into Postgres, and looking up index values will take longer. If you are expecting an extremely high volume of lookups, use Redis, otherwise, Postgres should suffice. 

I am not sure how much memory is required to run SonarSearch with Postgres, but it should not be a lot (2-4GB?). 

## Installation of tools 

Clone the SonarSearch git repository, and run the following commands: 

``` bash
make
make install
```

This will compile the various binaries used to set up the server and copy them to your path. You may wish to alter the install location specified in the make file. Or, you can omit the `make install` step and simply use the binaries from the `bin` directory after running `make`. 

Additionally, you will require either Postgres or Redis. You can use a Docker container for either of these, or run them locally. Consult google for setup instructions. 

The following command will spin up a Postgres container which can be used for the index: 
```bash 
docker run --name sonarsearch_postgres --expose 5432 -p 5432:5432 -v /var/lib/sonar_search:/var/lib/postgresql/data -e POSTGRES_PASSWORD=postgres -d postgres
```

## Set up Postgres 
Before you build the index, you must create the table in Postgres. This can be done with the following command: 
```bash 
psql -U postgres -h 127.0.0.1 -d postgres -c "CREATE TABLE crobat_index (id serial PRIMARY KEY, key text, value text)"
```

## Acquiring the datasets 

Dunno, good luck :) 

## Building the indexes  

To optimize searching these large datasets, a custom indexing strategy is used. Three steps are required in order to set this up: 

### Step 1 
First, you need to convert the project sonar dataset into the format used by SonarSearch. This can be done using the following command. 
``` bash
gunzip < 2021-12-31-1640909088-fdns_a.json.gz | sonar2crobat -i - -o crobat_unsorted
```

### Step 2 
In order to build the index, we need to sort the files obtained from the previous step. If you are running low on disk space, you can discard the raw gzip dataset. 

I recommend running these commands one at a time, as they are resource intensive: 

```
sort -k1,1 -k2,2 -t, crobat_unsorted_domains > crobat_sorted_domains
sort -k1,1 -t, -n crobat_unsorted_reverse > crobat_sorted_reverse
```

If you are happy, you can now discard the unsorted files.

### Step 3 
Once the files have been sorted, you need to generate indexes for both the subdomain and reverse DNS searches. 

To do so, you run the `crobat2index` binary, passing the input file, the format you wish to output (domain or reverse), and the storage backend (postgres or redis). 

`crobat2index` will output data to `stdout` which can be piped to either `redis-cli` or `psql` to import it quickly and efficiently. Below is an example of importing the `domain` index into Postgres.

```bash
crobat2index -i crobat_sorted_domains -f domain -backend postgres | psql -U postgres -h 127.0.0.1 -d postgres -c "COPY crobat_index(key, value) from stdin (Delimiter ',')"
```

Whereas inserting the `reverse` index would be done as follows: 
```bash
crobat2index -i crobat_sorted_reverse -f reverse -backend postgres | psql -U postgres -h 127.0.0.1 -d postgres -c "COPY crobat_index(key, value) from stdin (Delimiter ',')"
```

If something goes wrong and you need to try again, run this command: 
```bash
psql -U postgres -h 127.0.0.1 -d postgres -c "DROP TABLE crobat_index; CREATE TABLE crobat_index (id serial PRIMARY KEY, key text, value text)"
```

### Running crobat-server
Once you have completed all the previous steps, you are ready to run your crobat server. You will need to set a few env vars regarding configuration, as listed below:
```bash 
CROBAT_POSTGRES_URL=postgres://postgres:postgres@localhost:5432/postgres CROBAT_CACHE_BACKEND=postgres CROBAT_DOMAIN_FILE=~/Code/SonarSearch/testdata/crobat_sorted_domains CROBAT_REVERSE_FILE=~/Code/SonarSearch/testdata/crobat_sorted_reverse crobat-server
```

To make this easier to run, you can save these env variables to a file and source them. 

By default, `crobat-server` listens on ports 1997 (gRPC) and 1998 (HTTP).
### The end? 
You should now have a local working version of SonarSearch. Please note that postgres support is experimental, and may have some unexpected issues. If you encounter any problems, or have any questions regarding setup, feel free to open an issue on this repo. 
