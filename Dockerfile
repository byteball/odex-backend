FROM golang

# # Download gin for live reload (Usage: gin --path src --port 8081 run server.go serve)
# # RUN go get github.com/codegangsta/gin
# WORKDIR /go/src/app

RUN apt-get update
RUN apt-get install multitail

RUN mkdir -p /go/src/github.com/byteball/odex-backend
WORKDIR /go/src/github.com/byteball/odex-backend

ADD go.mod go.mod
ADD go.sum go.sum
RUN go mod download

COPY ./ ./

CMD ["go", "run", "main.go"]

EXPOSE 8081
EXPOSE 443
EXPOSE 80