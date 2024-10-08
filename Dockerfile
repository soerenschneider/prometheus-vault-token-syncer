FROM golang:1.23.2 AS build

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
ENV CGO_ENABLED=0
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags "-w" -o /prometheus-vault-token-syncer .


FROM gcr.io/distroless/static AS final

LABEL maintainer="soerenschneider"
USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /prometheus-vault-token-syncer /prometheus-vault-token-syncer

ENTRYPOINT ["/prometheus-vault-token-syncer"]
