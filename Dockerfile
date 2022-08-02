FROM golang:1.18 AS builder

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY . ./
RUN go build -o /docker-go-wr

FROM scratch
COPY --from=builder /docker-go-wr /docker-go-wr

EXPOSE 8080

CMD ["/docker-go-wr"]