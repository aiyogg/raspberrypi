#!/bin/bash

# DDNS for DNSPOD

oldIP=$(getent hosts pidns.tk | awk '{ print $1 }')
ip=$(curl -s https://api.ip.sb/ip)
if [[ -n $oldIP && -n $ip && $oldIP != $ip ]]; then
	res=$(curl -X POST https://dnsapi.cn/Record.Ddns -d 'login_token=123773,4ac6e09b1e3bc69ba1957ac5f25b9896&domain_id=76309514&record_id=481308506&sub_domain=@&record_line=默认&value=${ip}&format=json')
	echo -e "`date '+%c'`: \nSet DNS result:"
	echo $res
else
	echo -e "`date '+%c'`: \nOld IP: ${oldIP}"
	echo "Current IP: ${ip}"
fi
