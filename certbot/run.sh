#!/bin/bash

# ./first_run.sh
cp ./dns-cloudflare-renew.sh $HOME/CF-DDNS
if [[ -f "cloudflare.ini" ]]
then
    echo "Found cloudflare.ini file."
else
    echo "Please make sure cloudflare.ini file exists. Exiting..."
    exit 0
fi
cp ./cloudflare.ini $HOME/CF-DDNS
mkdir $HOME/CF-DDNS/cert
cp -rf ./cert/ $HOME/CF-DDNS/cert
crontab -l > cronlist
crontab -l | grep 'dns-cloudflare-renew.sh'
if [[ $? != 0 ]] ; then
    echo "0 18 1 * * /bin/bash $HOME/CF-DDNS/dns-cloudflare-renew.sh" >> cronlist
fi
crontab cronlist
rm cronlist