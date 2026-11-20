# Go 1.27.4: 2026-11-10 security release (12 CVEs incl. net/http).
FROM golang:1.27.4-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/albion-coach .

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/albion-coach /albion-coach
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/albion-coach"]
