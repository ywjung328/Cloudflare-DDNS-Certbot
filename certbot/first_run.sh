#!/bin/bash

docker run --rm -it --name certbot \
    -v ./cert/etc/letsencrypt:/etc/letsencrypt \
    -v ./cert/var/lib/letsencrypt:/var/lib/letsencrypt \
    -v ./cloudflare.ini:/cloudflare.ini \
    certbot/dns-cloudflare:latest \
    certonly \
    --dns-cloudflare \
    --dns-cloudflare-credentials /cloudflare.ini \
    -d jindocat.com \
    -d *.jindocat.com \
    --dns-cloudflare-propagation-seconds 60