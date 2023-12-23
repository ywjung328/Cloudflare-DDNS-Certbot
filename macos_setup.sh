#!/bin/bash

which -s go
if [[ $? != 0 ]] ; then
    which -s brew
    if [[ $? != 0 ]] ; then
        ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
    else
        echo "Brew is already installed. Updating brew now."
        brew update
    fi
    brew install go
else
    echo "Go is already installed. Skipping Go installation."
fi

echo "Building binary."
go build -o cf-ddns

if [[ -f "cf-ddns" ]]
then
    echo "Binary build done successfully."
else
    echo "There was problem while building binary. Exiting..."
    exit 0
fi

mkdir $HOME/CF-DDNS
cp ./cf-ddns $HOME/CF-DDNS

if [[ -f "config.json" ]]
then
    echo "Found config.json file."
else
    echo "Please make sure config.json file exists. Exiting..."
    exit 0
fi

cp ./config.json $HOME/CF-DDNS

echo "Adding cron job."
crontab -l > cronlist

crontab -l | grep '/cf-ddns'
if [[ $? != 0 ]] ; then
    echo "*/5 * * * * $HOME/CF-DDNS/cf-ddns" >> cronlist
fi

crontab -l | grep 'rm'
if [[ $? != 0 ]] ; then
    echo "* */12 * * * rm $HOME/CF-DDNS/logs/*.log" >> cronlist
fi

crontab cronlist
echo "Added cron job."

echo "Clearing directory."
rm cronlist
rm cf-ddns
echo "Clearing directory done."

echo "Test running the binary."
echo "Done. Logs can be found in ~/CF-DDNS/logs."
cd $HOME/CF-DDNS
./cf-ddns

echo "Setting done."