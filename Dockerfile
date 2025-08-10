# build and run application
FROM golang:1.24 AS build
    WORKDIR /app
    COPY . ./

    RUN go build


# install libreoffice
FROM ubuntu:latest AS libreoffice
    RUN apt-get update && \
        apt-get install -y libreoffice && \
        rm -rf /var/lib/apt/lists/*

    COPY --from=build /app/trile /app/trile

    CMD ["/app/trile", "--log-file=stdout"]
