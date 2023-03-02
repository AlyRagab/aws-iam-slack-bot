FROM golang:1.19.5-alpine as builder
WORKDIR /go/src
COPY . .
RUN go build -o iambot .

FROM alpine:3.17.2
WORKDIR /bin
COPY --from=builder /go/src .
USER nobody
EXPOSE 8080
CMD ["iambot"]
