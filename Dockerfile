FROM golang

RUN go get github.com/cgboal/sonarsearch/crobat

ENTRYPOINT ["crobat"]
