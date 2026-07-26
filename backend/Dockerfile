FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/immera-api ./cmd/api

FROM alpine:3.23
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
USER app
COPY --from=build /out/immera-api /usr/local/bin/immera-api
EXPOSE 8080
ENTRYPOINT ["immera-api"]
