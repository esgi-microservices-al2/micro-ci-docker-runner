FROM golang:latest as build

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM docker:dind as deploy
RUN mkdir /var/tar
WORKDIR /app
COPY --from=build /go/src/app/app .
CMD ["./app"]
