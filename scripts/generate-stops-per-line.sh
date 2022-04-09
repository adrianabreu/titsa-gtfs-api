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

tb push datasources/stops_per_line.datasource --force
tb push pipes/stops_per_line_fill.pipe --populate --force
tb push endpoints/query_stops_per_line.pipe --force