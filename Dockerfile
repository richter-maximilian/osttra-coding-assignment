FROM golang:1.20.2-alpine as build

WORKDIR /src

COPY . /src/

RUN go mod download

RUN CGO_ENABLED=0 go build -o main

FROM gcr.io/distroless/static:nonroot

USER 1000
COPY --from=build /src/main /main

COPY migrations /migrations
ENV DB_MIGRATIONS_DIR=file:///migrations

ENTRYPOINT ["/main"]
