FROM golang:alpine as builder

RUN apk add git
RUN go get github.com/golang/protobuf/proto
RUN go get google.golang.org/grpc
RUN git clone https://github.com/Cgboal/SonarSearch /go/src/github.com/Cgboal/SonarSearch
RUN cd /go/src/github.com/Cgboal/SonarSearch/crobat && export CGO_ENABLED=0 && go build -ldflags '-extldflags "-static"' && go install

FROM scratch
COPY --from=builder /go/bin/crobat /bin/crobat
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/bin/crobat"]
