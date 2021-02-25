FROM golang:1.15-alpine as build

WORKDIR /src
ADD go.mod .
ADD go.sum .
RUN go mod download
RUN apk add upx
ADD . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o igcds
RUN upx igcds

FROM scratch
COPY --from=build /src/igcds /bin/igcds
WORKDIR /run/config/cds
ENTRYPOINT [ "/bin/igcds", "-m", "groups", "-f", "cds.yaml.tpl", "-o", "cds.yaml" ]
