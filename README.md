# titsa-gtfs-api

This brief project exposes the current open data available from [Titsa](https://transitfeeds.com/p/transportes-interurbanos-de-tenerife/1058) (Bus Service on Tenerife)

And for getting that wonderful system working I'm using [tinybird](https://www.tinybird.co/)

If you register on tinybird the free tier would be enough. 

Advantages:
* It would expose the api quite soon (also with parameter support!) So we will have a quick api call - query in seconds
* It will be able to load the data directly
* The data could be processed before hand for diminishing data scanning.
* Its a wonderful experience.

![Free tier limtis](./imgs/free-tier-limits.png)

# 1.Loading data

If we go to the transitfeed url we will be able to download the last dataset available (in my case April1st). Also you can go to the oficcial titsa page and use this script I used for analysis: https://github.com/adrianabreu/titsa-gtfs-exploration/blob/master/download.sh

When we get the internal files we should rename then to csv, we will use everyone but the agency:

```
|
|_calendar_dates.csv
|_routes.csv
|_stop_times.csv
|_stops.csv
|_trips.csv
```

Then we need to register on tinybird (seriously the ui is so good I will skip it that part ) and start adding the files as data sources>

![loading data](./imgs/loading-data-source.png)

When we get all the data we will be able to generate a pipeline. THe pipelines are just like notebook cells and they can be easy parametizable. But this will be better done with an example.

I want an endpoint that given the current stop I'm intered in, the proper date and time, show me the next 5 five "guaguas" (buses for the non cuban - canarian people) incoming. 

So... Let's write the query, for the sake of avoiding transformations of parameters I'm considering that the callers is splitting the date time into two params as the calendar dates use the yyyymmdd and the stop times uses only the HH:MM:SS (so i don't have to deal with the +24 hours)

```
%
select routes.route_short_name as name, routes.route_long_name as headsign, stop_times.arrival_time as arrival
from routes 
inner join trips on routes.route_id = trips.route_id 
inner join calendar_dates on trips.service_id = calendar_dates.service_id 
inner join stop_times on stop_times.trip_id = trips.trip_id
where calendar_dates.date = {{Int64(calendar_date, 20220301, 'date of the service you are looking for as int YYYYMMDD', required=True)}}
AND stop_times.stop_id = {{Int64(stop_id, 1199, 'Stop id findable in titsa page', required=True)}} 
AND stop_times.arrival_time >= {{String(stop_time, '08:00:00', description='base search time', required=True)}}
order by arrival
limit 5
```

In the query there are two params defined using the template tinybird provides and we perform all the needed input params at once.

Now that is ready I want to try it, I will start downloading postman and then...
Oh wait, tinybird includes swagger! (Just click in view api on top right and then 'open in swagger' and the bottom)

![loading data](./imgs/swagger.png)

And here I'm looking for the buses for going to college are over again!

```
{
  "meta": [
    {
      "name": "name",
      "type": "Int32"
    },
    {
      "name": "headsign",
      "type": "String"
    },
    {
      "name": "arrival",
      "type": "String"
    }
  ],
  "data": [
    {
      "name": 105,
      "headsign": "SANTA CRUZ -> PUNTA DEL HIDALGO-POR LA LAGUNA",
      "arrival": "07:07:04"
    },
    {
      "name": 57,
      "headsign": "< > 51  CIRCULAR LA LAGUNA -> TEJINA - TACORONTE- LA LAGUNA",
      "arrival": "07:12:11"
    },
    {
      "name": 105,
      "headsign": "SANTA CRUZ -> PUNTA DEL HIDALGO-POR LA LAGUNA",
      "arrival": "07:17:53"
    },
    {
      "name": 50,
      "headsign": "LA LAGUNA -> TEGUESTE -BAJAMAR- PUNTA DEL HIDALGO",
      "arrival": "07:34:30"
    },
    {
      "name": 57,
      "headsign": "< > 51  CIRCULAR LA LAGUNA -> TEJINA - TACORONTE- LA LAGUNA",
      "arrival": "07:47:46"
    }
  ],
  "rows": 5,
  "rows_before_limit_at_least": 75,
  "statistics": {
    "elapsed": 0.085179922,
    "rows_read": 767212,
    "bytes_read": 21641662
  }
}
```