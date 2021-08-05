FROM golang:1.16 as build
LABEL maintainer "Shotaro Gotanda <g.sho1500@gmail.com>"

WORKDIR /work
COPY . /work
RUN --mount=type=cache,target=/go/pkg go build .

FROM ubuntu
COPY --from=build /work/k8stools /
ENTRYPOINT ["/k8stools"]

