package main

// GTFS data models
type Route struct {
	RouteID        string
	RouteShortName string
}

type Trip struct {
	RouteID   string
	ServiceID string
	TripID    string
}

type CalendarDate struct {
	ServiceID string
	Date      int
}

type StopTime struct {
	TripID        string
	ArrivalTime   string
	DepartureTime string
	StopID        int
	StopSequence  int
}

// Output record for Parquet
type TitsaArrival struct {
	Line             string `parquet:"name=line, type=BYTE_ARRAY, convertedtype=UTF8"`
	StopID           int32  `parquet:"name=stop_id, type=INT32"`
	EstimatedArrival int64  `parquet:"name=estimated_arrival, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
}
