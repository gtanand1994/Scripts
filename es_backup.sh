#!/bin/bash -x
# Backups Elastic search index into file in json format

gfile="/home/ubuntu/es_backup/getty-`date +%F`.json"

echo "$(date +%F): Backup local and getty under: $gfile and $lfile respectively" 

data=$(curl -L -vv 'http://<es_host>:9200/_cat/indices'| awk '{print $1" "$2" "$3" "$7" "$10}')
curl -X POST -H 'Content-type: application/json' --data "{'text':'$data'}" <SLACK_URL>;

/home/ubuntu/.nvm/versions/node/v8.10.0/bin/elasticdump --type=data --input=http://<es_host>:9200/<INDEX_NAME> --output=$gfile --limit=10 --maxSockets=1 2>&1 | tee >>~/es_backup.log

#Deleting backups older than 2 days
find /home/ubuntu/es_backup/ -type f -iname "getty*.json" -mtime +2 -exec rm -f {} \;
