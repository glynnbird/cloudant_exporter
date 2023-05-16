# Specifies a parent image
FROM golang:1.20.4 AS builder
 
# Creates an app directory to hold your app’s source code
WORKDIR /app
 
# Copies everything from your root directory into /app
COPY . .
 
# Installs Go dependencies
RUN go mod download
 
# Builds your app with optional configuration
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  ./cmd/couchmonitor
 
############################
# STEP 2 build a small image
############################
FROM alpine

# Copy our static executable.
COPY --from=builder /app/couchmonitor /
 
# Specifies the executable command that runs when the container starts
CMD [ "/couchmonitor", "--listen-address", "0.0.0.0:8080"]