# This image is a microservice in golang for the Degree chaincode
FROM golang:1.13.8-alpine AS build

WORKDIR /go/src/github.com/holzeis/lifecycle
ENV GO111MODULE=on

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

# Build application
RUN go build -o lifecycle .

# Production ready image
# Pass the binary to the prod image
FROM hyperledger/fabric-tools:2.0.1 as prod

COPY --from=build /go/src/github.com/holzeis/lifecycle/lifecycle /app/lifecycle

USER 1000

WORKDIR /app
CMD ./lifecycle

EXPOSE 8090