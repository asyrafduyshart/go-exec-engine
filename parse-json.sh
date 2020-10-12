#!/bin/bash

# Examples
# echo "{\"name\":\"John\", \"age\":31, \"city\":\"New York\"}" | ./parse-json.sh


# JSON read and turned into bash variables
read dataJson
eval "$(echo ${dataJson} | jq -r 'to_entries | .[] | .key + "=\"" + (.value|tostring) + "\""')"
echo ${name} ${age} ${city}
