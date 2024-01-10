###############
# Build stage #
###############
FROM golang:1.19 AS builder

ARG SOURCE_DIR

# Switch to /app dir
WORKDIR /app

# Cache dependencies download
COPY go.mod go.sum ./
RUN go mod download

# Copy the project files
COPY . .

ENV SOURCE_DIR=${SOURCE_DIR}

# Build the binary
RUN go build -o main ${SOURCE_DIR}/mock/*.go

# Set the binary as the CMD of the container
CMD ["./main"]

