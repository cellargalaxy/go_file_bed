FROM golang AS builder
ENV GOPROXY="https://goproxy.cn,direct"
ENV GO111MODULE=on
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /src/go-file-bed

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=builder /src/go-file-bed /go-file-bed
CMD ["/go-file-bed"]