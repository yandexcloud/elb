FROM golang:1.15-alpine as build

WORKDIR /src
ADD go.mod .
ADD go.sum .
RUN go mod download
RUN apk add upx
ADD *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w"
RUN upx igdns

FROM scratch
COPY --from=build /src/igdns /bin/igdns
WORKDIR /
ENTRYPOINT [ "/bin/igdns" ]
