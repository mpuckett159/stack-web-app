# Compile and export binary for Go app
FROM golang:1.16.3-buster AS build

ENV GO111MODULE=auto

WORKDIR /app
COPY . /app
RUN go build -o /app/main -ldflags="-extldflags=-static" -tags sqlite_omit_load_extension
CMD [ "/main" ]

# Copy out app from build image and set execution
FROM alpine:latest
COPY --from=build /app/main /
CMD [ "/main" ]
