FROM golang:1.20 AS build

ENV BASEDIR /go/src/github.com/MichielBijland/uncomplicated-registry

WORKDIR ${BASEDIR}

ADD . ${BASEDIR}

RUN go install -mod=vendor github.com/MichielBijland/uncomplicated-registry

FROM gcr.io/distroless/base:nonroot

COPY --from=build /go/bin/uncomplicated-registry /

ENTRYPOINT ["/uncomplicated-registry", "server"]
