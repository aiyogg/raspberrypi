#!/bin/bash

:<<!
Pull trackerlist from GitHub append to aria config file.
Use proxychains over GFW.
!

URL="https://raw.githubusercontent.com/ngosang/trackerslist/master/trackers_best.txt"
tracklist=$(proxychains3 curl -sfL "$URL")

list=$(echo $tracklist  | sed -e 's/ProxyChains-3\.1 (http:\/\/proxychains\.sf\.net)\s*//g'| sed -e 's/ /,/g')
echo $list

if [ -z $list ]; then
    echo "Request ${URL} failed."
    exit 0
fi
sed -i '$d' /home/pi/conf/aria2.conf
echo bt-tracker=$list >> /home/pi/conf/aria2.conf

containerid=$(docker ps -a | grep 'aria' | awk '{print $1}')
echo $containerid
if [[ -n $list && -n $containerid ]]; then
    docker restart $containerid
fi

echo "trackerlist refresh successful."
exit 0

