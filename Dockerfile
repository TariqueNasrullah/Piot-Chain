FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN apt-get update -y && apt-get install net-tools -y
RUN go get -d -v ./...
RUN go install
RUN cp /go/bin/iotchain /go/src/app/
