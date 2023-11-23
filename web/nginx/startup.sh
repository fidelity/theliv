#!/bin/bash

export DNS_RESOLVER=$(cat /etc/resolv.conf | grep nameserver | cut -d' ' -f2)
export EKS_DOMAIN=$(cat /etc/resolv.conf | grep search | cut -d' ' -f2)

# generate nginx.conf
export CERTS_PRIVATE=/nginx/theliv-private.pem
export CERTS_PUBLIC=/nginx/theliv-public.crt
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout ${CERTS_PRIVATE} -out ${CERTS_PUBLIC} -subj "/CN=theliv.io"

mkdir -p /nginx/conf.d/
mkdir -p /nginx/log/
mkdir -p /nginx/cache/
mkdir -p /nginx/run
mkdir -p /nginx/tmp
touch /nginx/run/nginx.pid

cp /etc/nginx/conf.d/default-temp.conf /nginx/conf.d/
cp /etc/nginx/nginx-temp.conf /nginx/
cp /etc/nginx/conf.d/datadog.conf /nginx/conf.d/

envsubst '$DNS_RESOLVER$CERTS_PRIVATE$CERTS_PUBLIC' </nginx/nginx-temp.conf > /nginx/nginx.conf

# generate default.con
envsubst '$EKS_DOMAIN$X_FORWARDED_PROTO$X_FORWARDED_HOST$ENVIRONMENT' </nginx/conf.d/default-temp.conf > /nginx/conf.d/default.conf
rm /nginx/conf.d/default-temp.conf

set -x
/app/server/main -ca "${ETCD_CA}" -key "${ETCD_KEY}" -cert "${ETCD_CERT}" -endpoints "${ETCD_ENDPOINTS}" & 
nginx -g 'daemon off;' -c /nginx/nginx.conf