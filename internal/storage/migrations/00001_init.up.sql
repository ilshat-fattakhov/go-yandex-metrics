BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS gaugemetrics ( 
	id SERIAL PRIMARY KEY,
	metricname VARCHAR (25) UNIQUE NOT NULL,
	metricvalue DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS gaugeidx
ON gaugemetrics(metricname);

CREATE TABLE IF NOT EXISTS countermetrics (
	id SERIAL PRIMARY KEY,
	metricname VARCHAR (25) UNIQUE NOT NULL,
	metricvalue BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS counteridx
ON countermetrics(metricname);

COMMIT;