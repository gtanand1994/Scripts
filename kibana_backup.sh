#!/usr/bin/env bash
set -x
usage="Usage: ${0} <elasticsearch url> <index>"
: ${1?:$usage}
: ${2?:$usage}

es=${1}
#initial=$(http -b $es/${2}/_search scroll==1m size==500 q=="@timestamp:>=now-15m")
initial=$(http -b $es/${2}/_search scroll==1m size==500)
scroll_id=$(echo $initial | jq -r '."_scroll_id"')
echo "Scroll id: ${scroll_id}"
hits=$(echo $initial | jq -r '.hits.hits | length')
echo $initial | jq '.hits.hits[]' > dump.json

until [[ $hits -eq 0 ]]
do
    results=$(http -b $es/_search/scroll scroll=1m scroll_id="${scroll_id}")
    echo $results | jq '.hits.hits[]' >> dump.json
    hits=$(echo $results | jq -r '.hits.hits | length')
done

http -b DELETE $es/_search/scroll scroll_id=$scroll_id

