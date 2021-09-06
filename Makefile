build:
	go build -o bin/sonar2crobat ./cmd/sonar2crobat
	go build -o bin/crobat2index ./cmd/crobat2index
	go build -tags=go_json -o bin/crobat-server ./cmd/crobat-server
	go build -o bin/crobat ./cmd/crobat

install:
	cp bin/* ${GOPATH}/bin/
