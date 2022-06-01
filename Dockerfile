FROM golang:1.17.8-alpine as builder
WORKDIR /go/src
COPY . .
RUN go build -o iambot .

FROM alpine:3.16.0
WORKDIR /bin
COPY --from=builder /go/src .
USER nobody
EXPOSE 8080
CMD ["iambot"]
