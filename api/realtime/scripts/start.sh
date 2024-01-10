#!/bin/bash

# ksqldb migrations config
MIGRATION_CONFIG="./ksqldb/ksql-migrations.properties"
JHOME="/usr/lib/jvm/java-11-openjdk"

export JAVA_HOME=${JHOME}
export PATH=$PATH:${JAVA_HOME}/jre/bin:${JAVA_HOME}/bin

abort() {
    echo "[ ABORT ] $1"
    exit 1
}

if [ -z ${KSQL_URL} ]; then
    abort "KSQL_URL env variable is not set"
fi

verify_ksqldb_is_available() {
    URL_TO_VERIFY=$1
    echo "checking if ${URL_TO_VERIFY} is available"
    ATTEMPTS=20
    for ATTEMPT in $(seq 1 $ATTEMPTS); do
        echo -n "attempt $ATTEMPT out of $ATTEMPTS... "
        curl --config ./curl_config.cfg ${URL_TO_VERIFY}
        if [ $? -eq 0 ]; then
            echo "connection succeeded"
            break
        else
            if [ $ATTEMPT -eq $ATTEMPTS ]; then
                abort "verification failed"
            else
                echo "retrying"
                sleep 3
            fi
        fi
    done
}

get_ksqldb_status_page() {
    DELAY=3
    URL_TO_VERIFY=$1

    echo "checking ${URL_TO_VERIFY}"
    ATTEMPTS=30
    OK_COUNTER=0

    for ATTEMPT in $(seq 1 $ATTEMPTS); do
        echo "attempt $ATTEMPT out of $ATTEMPTS... "
        echo "curl config file is $(readlink -f ./curl_config.cfg)"
        OUTPUT=$(curl --config ./curl_config.cfg ${URL_TO_VERIFY})

        if [ $? != 0 ]; then
            echo -n "curl exit code is non-zero, "
            if [ $ATTEMPT -eq $ATTEMPTS ]; then
                abort "status page verification failed"
            else
                echo "retrying in ${DELAY} seconds"
                sleep ${DELAY}
            fi
        fi

        echo "Response we have got:"
        echo ${OUTPUT} | jq .

        if [[ $(echo ${OUTPUT} | jq 'has("error_code")') == "true" ]]; then
            # for example
            # {
            #   "@type": "generic_error",
            #   "error_code": 50302,
            #   "message": "KSQL is not yet ready to serve requests."
            # }
            echo -n "error_code is present in output, "
            if [ $ATTEMPT -eq $ATTEMPTS ]; then
                abort "status page verification failed"
            else
                echo "retrying in ${DELAY} seconds"
                sleep ${DELAY}
            fi
        else
            # tmp solution, just to give ksqldb more time to bootstrap
            OK_COUNTER=$(expr $OK_COUNTER + 1)
            echo "status is ok, will proceed after 3 successful checks"
            if [ ${OK_COUNTER} -gt 3 ]; then break; fi
        fi

    done
}

import_custom_root_ca () {
    # custom root CA, to trust schema-registry TLS certificate
    CUSTOM_ROOT_CA_TMP_FILE="/tmp/tmp-root-ca.crt"

    echo -e "$1" > ${CUSTOM_ROOT_CA_TMP_FILE}

    sudo cp ${CUSTOM_ROOT_CA_TMP_FILE} /usr/local/share/ca-certificates/wme-internal-root-ca.crt
    sudo update-ca-certificates
    readlink -f $JAVA_HOME/jre/lib/security/cacerts
    sudo ln -sf /etc/ssl/certs/java/cacerts $JAVA_HOME/jre/lib/security/cacerts

    if [ -f ${CUSTOM_ROOT_CA_TMP_FILE} ]; then rm -f ${CUSTOM_ROOT_CA_TMP_FILE}; fi
}

run_ksql_migrations () {
    # ksqldb migrations log dir
    export LOG_DIR="/home/$(whoami)/ksql_migration_logs"

    echo "[ ENTRY ] ksql-migrations -c $1 initialize-metadata"
    INIT_METADATA_RES=$(ksql-migrations -c "$1" initialize-metadata 2>&1)
    if [ $? -ne 0 ]; then
        if [[ ${INIT_METADATA_RES} == *"A stream with the same name already exists"* ]]; then
            echo "initialize-metadata: we've received an expected error, proceeding: A stream with the same name already exists"
        else
            abort "there were issues during running initialize-metadata: ${INIT_METADATA_RES}"
        fi
    fi

    echo "[ ENTRY ] ksql-migrations INFO, before APPLY"
    ksql-migrations -c "$1" info

    echo "[ ENTRY ] ksql-migrations APPLY"
    ksql-migrations -c "$1" apply --all || abort "ksql-migration APPLY has failed"

    echo "[ ENTRY ] ksql-migrations INFO, after APPLY"
    ksql-migrations -c "$1" info

    echo "[ END ] ksql-migrations finished"

}

prepare_curl_config() {
    WITH_AUTH=$1
    CONFIG_LINE="-s\n--max-time 2"
    if [ "${WITH_AUTH}" == "with_auth" ]; then
        echo "[ INFO ] adding creds to curl config"
        CONFIG_LINE="${CONFIG_LINE}\n--user ${KSQL_USERNAME}:${KSQL_PASSWORD}"
    fi
    echo -e "${CONFIG_LINE}" > ./curl_config.cfg
    echo "generated $(readlink -f ./curl_config.cfg)"
}

run_main_app() {
    # run main app
    echo "[ ENTRY ] Running main app"
    ./main
}

export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
export SSL_CERT_DIR=/etc/ssl/certs

if [ -z "${INTERNAL_ROOT_CA_PEM}" ]; then
    echo "INTERNAL_ROOT_CA_PEM is not set"
else
    import_custom_root_ca "${INTERNAL_ROOT_CA_PEM}"
fi

# check if KSQL_URL  ends with /, we don't want // in logs.
# also looks like // makes ksqldb return an empty result
if [[ "$KSQL_URL" == */ ]]; then
    KSQL_STATUS_PAGE=${KSQL_URL}status
else
    KSQL_STATUS_PAGE=${KSQL_URL}/status
fi

# here we are preparing parameters line for curl.
# in case auth is used - same file will containe the needed creds.
if [ -z "$KSQL_USERNAME" ] || [ -z "$KSQL_PASSWORD" ] ; then
    echo "[ WARN ] We did not find auth credentials in environment, so proceeding without authentication"
    prepare_curl_config
else
    echo "[ INFO ] Auth creds found, activating auth config"
    prepare_curl_config "with_auth"
    # also here, in one shot we will add auth username and password to
    # migrations tool config
    sudo sed -i "s/#\ ksql.auth.basic.username=/ksql.auth.basic.username=${KSQL_USERNAME}/" ${MIGRATION_CONFIG}
    sudo sed -i "s/#\ ksql.auth.basic.password=/ksql.auth.basic.password=${KSQL_PASSWORD}/" ${MIGRATION_CONFIG}
fi

# if auth is used - we should have ready curl config already at this point
verify_ksqldb_is_available $KSQL_URL
get_ksqldb_status_page $KSQL_STATUS_PAGE

# fault prevention measure - we will skip migration in case KSQL_MIGRATIONS_ENABLED env var
# is set to (F|f)alse, for example via IAC/taskdef
if [[ "${KSQL_MIGRATIONS_ENABLED}" == "false" || "${KSQL_MIGRATIONS_ENABLED}" == "False" ]]; then
    echo "KSQL_MIGRATIONS_ENABLED is set to False, skipping run_ksql_migrations"
else
    echo "Running ksqldb migrations (set KSQL_MIGRATIONS_ENABLED to (F|f)alse to disable)"
    sudo sed -i "s#___KSQL_URL___#${KSQL_URL}#g" ${MIGRATION_CONFIG}
    run_ksql_migrations ${MIGRATION_CONFIG}
fi

run_main_app
