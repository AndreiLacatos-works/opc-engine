# build stage
FROM golang:1.23-alpine AS builder

# ensure make is installed
RUN apk add --no-cache make

WORKDIR /build

# copy go dep files & install libs
COPY src/go.* ./
RUN go mod download

# copy source
WORKDIR /build/src
COPY src/ /build/src/

# build the app
RUN make build-release


# package stage
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /build/src/release/opc-engine-simulator /app/opc-engine-simulator

ENTRYPOINT [ "/app/opc-engine-simulator" ]


# BELOW STUFF WORKS
# FROM golang:1.23

# WORKDIR /app

# COPY src/go* ./

# RUN go mod download

# COPY src/ /app/src/

# WORKDIR /app/src

# RUN make build-release

# ENTRYPOINT [ "./release/opc-engine-simulator" ]








# BROKEN FOR WHATEVER REASON
# FROM golang:1.23 AS builder

# # set the working directory in the container
# WORKDIR /app

# # copy the entire source folder into the container
# COPY src/ /app/src/

# # install make
# RUN apt-get update && apt-get install -y make

# # build the project
# WORKDIR /app/src/
# RUN make build-release

# # use alpine as base image
# FROM alpine:3.21

# # set the working directory in the second image
# WORKDIR /root/

# # copy the binary from the builder image
# COPY --from=builder /app/src/release/opc-engine-simulator .

# # run the application
# CMD ["./opc-engine-simulator"]