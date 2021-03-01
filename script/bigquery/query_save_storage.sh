#!/bin/sh
echo "Querying data from big query"
echo ">>>" "$query"
echo ">>>" $storageLocation
# SELECT * FROM `lido-white-label.kafka_production.transaction_notification` WHERE DATE(_PARTITIONTIME) = "2021-02-25" LIMIT 1

sendPubNub(){
    curl --silent --output /dev/null --request POST --url http://localhost:7001/pubnub/publish/$1 --header 'Content-Type: application/json' --data '{ "message": '\""$2\""' }'
}

runBigQuery() {
    sendPubNub 'exec_engine.alert' 'Executing Big Query'
    timestamp=$(date +%s)
    # echo "execute command: bq query --use_legacy_sql=false "EXPORT DATA OPTIONS(uri='$storageLocation', format='JSON') AS $query""
    echo "result:"
    echo "\n"
    if [ -z "$storageLocation" ]
    then
        echo "\$storageLocation is empty, proceed with default uri"
        uri="gs://lido-white-label-data/production/test-avro/backup-$timestamp-*"
    else
        echo "\$storageLocation is NOT empty"
        uri=$storageLocation
    fi
    bq query --use_legacy_sql=false "EXPORT DATA OPTIONS(uri='$uri',  format='JSON', compression='GZIP', overwrite=true) AS $query" >> ./script/bigquery/log/result.log
    if [ $? -eq 0 ]; then
        sendPubNub 'exec_engine.alert.success' 'BigQuery Execution Success'
        echo '--- BigQuery Execution Success'
    else
        sendPubNub 'exec_engine.alert.failed' 'BigQuery Execution Failed'
        echo '--- BigQuery Execution Failed'
    fi
}

runBigQuery
