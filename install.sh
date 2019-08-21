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

wget 'https://github.com/cellargalaxy/goFileBed/releases/download/v0.0.2/goFileBed-linux'
chmod 755 ./goFileBed-linux
./goFileBed-linux

echo $config > $configPath

docker build -t go_file_bed .
docker run -d -v file_bed:/file_bed