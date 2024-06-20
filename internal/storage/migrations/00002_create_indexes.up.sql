BEGIN TRANSACTION;

CREATE INDEX IF NOT EXISTS gaugeidx
ON gaugemetrics(metricname);

CREATE INDEX IF NOT EXISTS counteridx
ON countermetrics(metricname);

COMMIT;