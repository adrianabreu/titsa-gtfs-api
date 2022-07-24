TITSA_URL=http://apps.titsa.com/apps/apps_sae_llegadas_parada.asp
STOP_ID=1918

ANSWER="$(curl -i --no-progress-meter -X GET "$TITSA_URL?idApp=$TITSA_TOKEN&idParada=$STOP_ID")"
STATUS=$(echo "$ANSWER" | grep HTTP |  awk '{print $2}')
STATUSCASTED=$(($STATUS + 0))
BODY=$(echo "$ANSWER" | grep "<?xml") 
if [ $STATUSCASTED -eq 200 ]
then

DATETIME="$(xmllint --xpath "/llegadas[1]/llegada/hora/text()" - <<<"$BODY")"
CALENDARDATE="$(cut -d ' ' -f1 <<< $DATETIME)"
ARRIVALHOUR="$(cut -d ' ' -f2 <<< $DATETIME)"
IFS=/ read day month year <<< "$CALENDARDATE"
IFS=: read hour min sec <<< "$ARRIVALHOUR"
next_mins=$(xmllint --xpath "/llegadas[1]/llegada/minutosParaLlegar/text()" - <<<"$BODY")
echo $next_mins
new_mins="$(($min + $next_mins))"
if [ $new_mins -gt 60 ]
then
hour="$(($hour + 1))"
new_mins="$(($new_mins - 60))"

fi

printf -v hour "%02d" $hour
printf -v new_mins "%02d" $new_mins

REQUEST="{
    \"line\":\"$(xmllint --xpath "/llegadas[1]/llegada/linea/text()" - <<<"$BODY")\",
    \"stop_id\":\"$(xmllint --xpath "/llegadas[1]/llegada/codigoParada/text()" - <<<"$BODY")\",
    \"calendar_date\": \"${year}${month}${day}\",
    \"arrival_time\": \"${hour}:${new_mins}:${sec}\"
}"

    curl \
    -H "Authorization: Bearer $TINYBIRD_TOKEN" \
    -d "$REQUEST" \
    'https://api.tinybird.co/v0/datasources?format=ndjson&name=realtime&mode=append' \

else
    exit "$STATUS"
fi
