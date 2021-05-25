# Compile and export binary for Go app
FROM golang:1.16.3-buster AS build

ENV GO111MODULE=auto

WORKDIR /app
COPY . /app
RUN go build -o /app/main

# Copy out app from build image and set execution
FROM alpine:latest
RUN apk add libc6-compat

COPY --from=build /app/main /

CMD [ "/main" ]
