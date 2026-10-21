FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /openai_cost_exporter ./cmd/openai_cost_exporter

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /openai_cost_exporter /openai_cost_exporter
USER 65532:65532
EXPOSE 9464
ENTRYPOINT ["/openai_cost_exporter"]
CMD ["--web.listen-address=:9464"]
