#!/usr/bin/env bash
configPath="config.yml"

if [ ! -f $configPath ]; then
    echo 'config not exist'
    while :
    do
        read -p "please enter token(required):" token
        if [ ! -z $token ];then
            break
        fi
    done
    read -p "please enter synUrl(default:http://127.0.0.1:8880):" synUrl
    if [ -z $synUrl ];then
        synUrl="http://127.0.0.1:8880"
    fi
    read -p "please enter listeningAddress(default:0.0.0.0:8880):" listeningAddress
    if [ -z $listeningAddress ];then
        listeningAddress="0.0.0.0:8880"
    fi
    config='token: '$token'
synUrl: '$synUrl'
listeningAddress: '$listeningAddress'
fileBedPath: file_bed'
else
    echo 'config exist'
    config=`cat $configPath`
fi

echo $config
echo 'config,input any key go on,or control+c over'
read
echo $config > $configPath