FROM golang AS builder
ENV GOPROXY="https://goproxy.io"
ENV GO111MODULE=on
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /src/goFileBed

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=builder /src/goFileBed /goFileBed
CMD ["/goFileBed"]