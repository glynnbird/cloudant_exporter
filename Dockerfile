# Specifies a parent image
FROM golang:1.23.6 AS builder
 
# Creates an app directory to hold your app’s source code
WORKDIR /app
 
# Copies everything from your root directory into /app
COPY . .
 
# Installs Go dependencies
RUN go mod download
 
# Builds your app with optional configuration
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build
 
############################
# STEP 2 build a small image
############################
# Development
FROM alpine
# Production
# FROM gcr.io/distroless/static-debian11

# Copy our static executable.
COPY --from=builder /app/cloudant_exporter /
 
# Specifies the executable command that runs when the container starts
CMD [ "/cloudant_exporter", "--listen-address", "0.0.0.0:8080"]
