FROM golang AS builder
ENV GOPROXY="https://goproxy.cn,direct"
ENV GO111MODULE=on
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix -o /src/go_file_bed .

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=builder /src/go_file_bed /go_file_bed
CMD ["/go_file_bed"]