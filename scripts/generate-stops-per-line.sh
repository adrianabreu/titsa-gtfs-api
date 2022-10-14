cat > ./datasources/stops_per_line.datasource << 'EOF'
SCHEMA >
    `line` Int32,
    `stop_id` Int32

ENGINE "AggregatingMergeTree"
ENGINE_SORTING_KEY line, stop_id 
EOF

cat > ./pipes/stops_per_line_fill.pipe << 'EOF'
NODE materialization
DESCRIPTION >
    Materializes results into stops_per_line
SQL >
    SELECT routes.route_short_name as line, stop_times.stop_id as stop_id
    FROM stop_times 
    INNER JOIN trips on stop_times.trip_id = trips.trip_id
    INNER JOIN routes on routes.route_id = trips.route_id
    GROUP BY routes.route_short_name as line,  stop_times.stop_id as stop_id

TYPE materialized
DATASOURCE stops_per_line
EOF

cat > ./endpoints/query_stops_per_line.pipe << 'EOF'
TOKEN public_read_token READ

NODE query
DESCRIPTION >
    Query all stops per line
SQL >
    %
    SELECT stop_id
    FROM stops_per_line 
    WHERE line = {{Int32(line, required=True)}}
EOF

cat > ./endpoints/realtime_vs_predicted.pipe << 'EOF'
NODE realtime_vs_predicted_0
SQL >

    %
    select distinct (*)
    from
        (
            (
                select
                    routes.route_short_name as line,
                    toInt32(stop_times.stop_id) as stop_id,
                    toString(calendar_dates.date) as calendar_date,
                    stop_times.arrival_time as arrival_time,
                    0 as is_realtime
                from routes
                inner join trips on routes.route_id = trips.route_id
                inner join
                    calendar_dates on trips.service_id = calendar_dates.service_id
                inner join stop_times on stop_times.trip_id = trips.trip_id
                where
                    stop_times.stop_id = 1918
                    and calendar_dates.date >= {{ Int32(start_date, 20220926) }}
                    and calendar_dates.date <= {{ Int32(end_date, 20220927) }}
            )
            union all
            select *, 1 as is_realtime
            from realtime
            where
                calendar_date >= {{ String(start_date, "20220926") }}
                and calendar_date <= {{ String(end_date, "20220927") }}
        )
    order by calendar_date asc, arrival_time asc



NODE realtime_vs_predicted_1
SQL >

    with parsed_time as (
    SELECT
        *,
        toDateTime(concat(
            substring(calendar_date, 1, 4),
            '-',
            substring(calendar_date, 5, 2),
            '-',
            substring(calendar_date, 7, 2),
            ' ',
            arrival_time
        )) as time
    FROM realtime_vs_predicted_0
    )
    select *,
        if(neighbor(is_realtime, -1) = 1, neighbor(time, -1), null) AS prev,
        if(neighbor(is_realtime, 1) = 1, neighbor(time, 1), null) AS next
    from parsed_time
    order by time asc



NODE realtime_vs_predicted_2
SQL >

    select
        concat(
            substring(calendar_date, 1, 4),
            '-',
            substring(calendar_date, 5, 2),
            '-',
            substring(calendar_date, 7, 2)
        ) as date,
        sum(
            if(next is not null, if((next - time) > (20 * 60), 0, next - time), 0) as t
        ) as acc_diff_in_seconds
    from realtime_vs_predicted_1
    where (prev is not null or next is not null) and is_realtime = 0
    group by date

EOF

tb push datasources/stops_per_line.datasource --force
tb push pipes/stops_per_line_fill.pipe --populate --force
tb push endpoints/query_stops_per_line.pipe --force
tb push endpoints/realtime_vs_predicted.pipe --force


