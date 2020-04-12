#!/usr/bin/env bash

while :
do
    if [ ! -z $token ];then
        break
    fi
    read -s -p "please enter token(required):" token
done

if [ -z $listenPort ];then
    read -p "please enter listen port(default:8880):" listenPort
fi
if [ -z $listenPort ];then
    listenPort="8880"
fi

if [ -z $lastFileInfoCount ];then
    read -p "please enter last file info count(default:10):" lastFileInfoCount
fi
if [ -z $lastFileInfoCount ];then
    lastFileInfoCount="10"
fi

if [ -z $pullSyncCron ];then
    read -p "please enter pullSyncCron(default:''):" pullSyncCron
fi
if [ -z $pullSyncHost ];then
    read -p "please enter pullSyncHost(default:''):" pullSyncHost
fi
if [ -z $pullSyncToken ];then
    read -p "please enter pullSyncToken(default:''):" pullSyncToken
fi
if [ -z $pushSyncCron ];then
    read -p "please enter pushSyncCron(default:''):" pushSyncCron
fi
if [ -z $pushSyncHost ];then
    read -p "please enter pushSyncHost(default:''):" pushSyncHost
fi
if [ -z $pushSyncToken ];then
    read -p "please enter pushSyncToken(default:''):" pushSyncToken
fi

echo 'token:'$token
echo 'listenPort:'$listenPort
echo 'lastFileInfoCount:'$lastFileInfoCount
echo 'pullSyncCron:'$pullSyncCron
echo 'pullSyncHost:'$pullSyncHost
echo 'pullSyncToken:'$pullSyncToken
echo 'pushSyncCron:'$pushSyncCron
echo 'pushSyncHost:'$pushSyncHost
echo 'pushSyncToken:'$pushSyncToken
echo 'input any key go on,or control+c over'
read

echo 'docker build'
docker build -t go_file_bed .
echo 'docker create volume'
docker volume create file_bed
echo 'docker run'
docker run -d \
--restart=always \
--name go_file_bed \
-p $listenPort:8880 \
-e TOKEN=$token \
-e LAST_FILE_INFO_COUNT=$lastFileInfoCount \
-e PULL_SYNC_CRON=$pullSyncCron \
-e PULL_SYNC_HOST=$pullSyncHost \
-e PULL_SYNC_TOKEN=$pullSyncToken \
-e PUSH_SYNC_CRON=$pushSyncCron \
-e PUSH_SYNC_HOST=$pushSyncHost \
-e PUSH_SYNC_TOKEN=$pushSyncToken \
-v file_bed:/file_bed \
go_file_bed

echo 'all finish'