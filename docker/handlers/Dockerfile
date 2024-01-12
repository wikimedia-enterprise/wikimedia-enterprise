###############
# Build stage #
###############
FROM golang:1.19 AS builder

# source to build as an argument, for shared CI/CD
ARG SOURCE_DIR=null

# Switch to /app dir
WORKDIR /app

# Cache dependencies download
COPY go.mod go.sum ./
RUN go mod download

# Copy the project files
COPY . .

# Build the binary
RUN go build -o main ${SOURCE_DIR}/*.go

#############
# Run stage #
#############
FROM alpine:latest

# Set name of app user
ARG USERNAME=gorunner

# Install glibc compatibility
RUN apk --no-cache add gcompat ca-certificates sudo && rm -rf /var/cache/apk/*

# Add app user
RUN adduser -u 1000 -S ${USERNAME} -G wheel -D alpine

# sudo - allow cp and update-ca-certificates, for wheel group users. Needed for custom truststores
RUN echo "%wheel ALL=(root) NOPASSWD:$(which cp), $(which update-ca-certificates)" > /etc/sudoers.d/wheel

# Use app user
USER ${USERNAME}

# Switch to working directory
WORKDIR /home/${USERNAME}

# Copy binary from previous stage
COPY --from=builder /app/main ./

# Copy startup script which will dynamically add custom truststore if INTERNAL_ROOT_CA_PEM env var is set
COPY ./docker/handlers/start.sh ./

# Set the binary as the CMD of the container
CMD ["./start.sh"]
