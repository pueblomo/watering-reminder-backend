FROM golang:alpine AS builder

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
ADD go.mod .
ADD go.sum .

COPY . .
RUN go build -o wr main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/wr /app/wr

EXPOSE 8080

CMD ["./wr"]