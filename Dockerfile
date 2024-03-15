FROM golang:1.21.6-bullseye as builder

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY go.* ./
ENV GOPRIVATE=github.com/firehydrant/*
RUN --mount=type=secret,id=netrc,target=/root/.netrc go mod download -x

ADD ./ /go/src/app/

ARG REVISION ${REVISION:-unknown}

RUN go build -o signals-migrator -ldflags="-X 'main.Revision=${REVISION}'" main.go

FROM debian:bullseye as release

RUN apt-get update && apt-get install -y \
  ca-certificates\
  postgresql-client\
  && apt-get clean

RUN addgroup firehydrant && adduser -u 1000 --ingroup firehydrant firehydrant
USER 1000

COPY --from=builder /go/src/app/signals-migrator /usr/local/bin/signals-migrator

ENTRYPOINT ["/usr/local/bin/signals-migrator"]
