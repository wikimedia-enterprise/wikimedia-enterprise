ARG KSQLDB_RELEASE="0.28.2"
ARG KSQLDB_DIR="/usr/share/java/ksqldb-rest-app"
ARG USERNAME="gorunner"

###############
# Build stage #
###############
FROM golang:1.19 AS builder

# Switch to /app dir
WORKDIR /app

# Cache dependencies download
COPY go.mod go.sum ./
RUN go mod download

# Copy the project files
COPY . .

# Build the binary
RUN go build -o main *.go

################
# ksqldb stage #
################
# we are using this stage only to copy ksql-migrations stuff
FROM confluentinc/ksqldb-server:${KSQLDB_RELEASE} as ksqldb
ARG KSQLDB_DIR
RUN ls -alrth ${KSQLDB_DIR}

#############
# Run stage #
#############
FROM alpine:latest

# inherit global values
ARG KSQLDB_RELEASE
ARG KSQLDB_DIR
ARG USERNAME

# Copy ksql-migrations from ksqldb image
# we are not copying everything to save around 130Mb of space
RUN mkdir -p ${KSQLDB_DIR}
RUN mkdir -p /usr/config
COPY --from=ksqldb /usr/bin/ksql-run-class /usr/bin/ksql-run-class
COPY --from=ksqldb /usr/bin/ksql-migrations /usr/bin/ksql-migrations
COPY --from=ksqldb /etc/ksqldb/log4j.properties /etc/ksqldb/log4j.properties
COPY --from=ksqldb ${KSQLDB_DIR}/* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/ksqldb-tools-${KSQLDB_RELEASE}.jar ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/slf4j* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/airline* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/common* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/javax.inject* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/ksqldb* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/kafka* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/*log4j* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/guava* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/vertx* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/netty* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/jackson* ${KSQLDB_DIR}/
# COPY --from=ksqldb ${KSQLDB_DIR}/antlr* ${KSQLDB_DIR}/
RUN du -h ${KSQLDB_DIR}

# copy log4j config which should redirect logs to stdout
COPY ./submodules/cdtools/docker/ksqldb_services/log4j.properties /usr/config/ksql-migrations-log4j.properties

# Install glibc compatibility, jdk
RUN apk --no-cache add gcompat ca-certificates sudo bash openjdk11-jre java-cacerts curl jq && rm -rf /var/cache/apk/*

# Add app user
RUN adduser -u 1000 -S ${USERNAME} -G wheel -D alpine

# sudo - allow cp and update-ca-certificates, for wheel group users. Needed for custom truststores
RUN echo "%wheel ALL=(root) NOPASSWD:$(which cp), $(which update-ca-certificates), $(which ln), $(which sed), $(which tee)" > /etc/sudoers.d/wheel

# ensure latest curl is installed
RUN apk update curl

# Use app user
USER ${USERNAME}

# Switch to working directory
WORKDIR /home/${USERNAME}
RUN mkdir -p ksql_migration_logs

# Copy binary from previous stage
COPY --from=builder /app/main ./

# Copy ksqldb migrations
COPY ./ksqldb/ ./ksqldb

# debug step
# RUN ksql-migrations -c ./ksqldb/ksql-migrations.properties initialize-metadata

# Copy startup script which will dynamically add custom truststore if INTERNAL_ROOT_CA_PEM env var is set
COPY ./submodules/cdtools/docker/ksqldb_services/start.sh ./

# Set the binary as the CMD of the container
CMD ["./start.sh"]
