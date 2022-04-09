#!/bin/bash
wget http://www.titsa.com/Google_transit.zip
rm -rf ./tmp_data
unzip Google_transit.zip -d ./tmp_data
rm Google_transit.zip

FILES="./tmp_data/*"
for f in $FILES
do
  tb datasource generate "$f" --force
  tb push --force datasources/"$(basename -- $f .txt)".datasource
done