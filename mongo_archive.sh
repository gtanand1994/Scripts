#!/bin/bash

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <retention days> " >&2
  exit 1
fi

n=$1
S3_BUCKET="mongo-archive-blowhorn"
pass=`cat /home/ubuntu/scripts/.mongopass`
mongo_host="mongodb://blowhorn-prod-shard-00-00-jsdo3.mongodb.net:27017,blowhorn-prod-shard-00-01-jsdo3.mongodb.net:27017,blowhorn-prod-shard-00-02-jsdo3.mongodb.net:27017/test?replicaSet=Blowhorn-Prod-shard-0"
mongodump_host="Blowhorn-Prod-shard-0/blowhorn-prod-shard-00-00-jsdo3.mongodb.net:27017,blowhorn-prod-shard-00-01-jsdo3.mongodb.net:27017,blowhorn-prod-shard-00-02-jsdo3.mongodb.net:27017"
lastMon=`echo -e "var lastMonth = new Date() \n lastMonth.setDate(lastMonth.getDate() - $n) \n lastMonth.getTime()" | mongo $mongo_host --quiet --ssl --authenticationDatabase admin --username bhmongo --password $pass |tail -1`
if echo $lastMon |wc -c |grep -vq 14;
then
	echo "ERROR: Issue with creating retention time using mongo query, exiting"
	exit
fi
h_lastMon=`echo -e "new Date($lastMon).toISOString()" | mongo $mongo_host --quiet --ssl --authenticationDatabase admin --username bhmongo --password $pass |tail -1`

echo "1. Backing up data older than: $h_lastMon"
k=`echo $h_lastMon|cut -d "T" -f 1`

dir="/home/ubuntu/scripts/`date --iso-8601=seconds | cut -d '+' -f 1|sed 's/://g'|sed 's/T/_'${k}'T/g'`"
echo $dir

echo "====================Backing up activity_driverofflinelog Collection===================="
mongodump --host $mongodump_host -o $dir --ssl --username bhmongo --password $pass --authenticationDatabase admin --db aquarius -c 'activity_driverofflinelog' -q '{"createdTime": {"$lte" : new Date('$lastMon')}}' --gzip
if echo $?|grep -q 0; then echo "INFO: OK, Dump ran successfully"; else echo "ERROR: MongoDump exited abnormally";exit; fi

echo "====================Backing up activity_drivertriplog Collection===================="
mongodump --host $mongodump_host -o $dir --ssl --username bhmongo --password $pass --authenticationDatabase admin --db aquarius -c 'activity_drivertriplog' -q '{"createdTime": {"$lte" : new Date('$lastMon')}}' --gzip
if echo $?|grep -q 0; then echo "INFO: OK, Dump ran successfully"; else echo "ERROR: MongoDump exited abnormally";exit; fi

echo "2. Compressing the backup"
tar -vc $dir |gzip > ${dir}.tar.gz

echo "3. Verifying the taken backup: ${dir}.tar.gz"
if tar -tzvvf ${dir}.tar.gz;
then
	echo "INFO: OK, Backup looks good";
else
	echo "ERROR: Backup is corrupted, So not proceeding further";
	exit
fi
rm -rf $dir
mv ${dir}.tar.gz /home/ubuntu/scripts/s3_mongo_backups/

echo "3. Verifying the S3 Bucket: $S3_BUCKET :"
if aws s3 ls "s3://$S3_BUCKET" 2>&1 | grep -q 'NoSuchBucket'
then
	echo "DEBUG: $S3_BUCKET doesn't exist, so creating it"
	aws s3 mb s3://$S3_BUCKET
else
	echo "INFO: OK, Bucket is present $S3_BUCKET"
fi

cd /home/ubuntu/scripts/s3_mongo_backups/
echo "4. Keeping only 3 backups under /home/ubuntu/scripts/s3_mongo_backups/"
ls -t /home/ubuntu/scripts/s3_mongo_backups/ |awk 'NR>3'
rm $(ls -t /home/ubuntu/scripts/s3_mongo_backups/ |awk 'NR>3')

echo "5. Syncing Local backup folder /home/ubuntu/scripts/s3_mongo_backups/ with remote S3 Bucket $S3_BUCKET"
aws s3 sync /home/ubuntu/scripts/s3_mongo_backups/ s3://$S3_BUCKET/
echo "INFO: OK, S3 Bucket $S3_BUCKET is updated"
aws s3 ls "s3://$S3_BUCKET"

## Deleting old data from the corresponding collections
#echo "6. Truncating $n days older data from activity_driverofflinelog and activity_drivertriplog"
#echo -e "use aquarius \n db.activity_driverofflinelog.remove({\"createdTime\": {\"$lte\" : new Date('$lastMon')}}).sort({_id:-1})" | mongo $mongo_host --quiet --ssl --authenticationDatabase admin --username bhmongo --password $pass
#if echo $?|grep -q 0; then echo "INFO: OK, Truncated activity_driverofflinelog collection successfully"; else echo "ERROR: While Truncating activity_driverofflinelog collection successfully, exiting";exit; fi
#echo -e "use aquarius \n db.activity_drivertriplog.remove({\"createdTime\": {\"$lte\" : new Date('$lastMon')}}).sort({_id:-1})" | mongo $mongo_host --quiet --ssl --authenticationDatabase admin --username bhmongo --password $pass
#if echo $?|grep -q 0; then echo "INFO: OK, Truncated activity_drivertriplog collection successfully"; else echo "ERROR: While Truncating activity_driverofflinelog collection successfully, exiting";exit; fi
