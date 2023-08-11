#!/bin/bash

export DNS_RESOLVER=$(cat /etc/resolv.conf | grep nameserver | cut -d' ' -f2)
export EKS_DOMAIN=$(cat /etc/resolv.conf | grep search | cut -d' ' -f2)

# generate nginx.conf
export CERTS_PRIVATE=/etc/ssl/certs/theliv-private.pem
export CERTS_PUBLIC=/etc/ssl/certs/theliv-public.crt
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout ${CERTS_PRIVATE} -out ${CERTS_PUBLIC} -subj "/CN=theliv.io"
envsubst '$DNS_RESOLVER$CERTS_PRIVATE$CERTS_PUBLIC' </etc/nginx/nginx-temp.conf > /etc/nginx/nginx.conf

# generate default.con
envsubst '$EKS_DOMAIN$X_FORWARDED_PROTO$X_FORWARDED_HOST$ENVIRONMENT' </etc/nginx/conf.d/default-temp.conf > /etc/nginx/conf.d/default.conf
rm /etc/nginx/conf.d/default-temp.conf

set -x
/app/server/main -ca "${ETCD_CA}" -key "${ETCD_KEY}" -cert "${ETCD_CERT}" -endpoints "${ETCD_ENDPOINTS}" & 
nginx -g 'daemon off;'