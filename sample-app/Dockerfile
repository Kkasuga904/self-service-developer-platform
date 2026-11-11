FROM golang:1.24.6-alpine3.22@sha256:c8c5f95d64aa79b6547f3b626eb84b16a7ce18a139e3e9ca19a8c078b85ba80d AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sample-app .

FROM scratch
COPY --from=build /out/sample-app /sample-app
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/sample-app"]
