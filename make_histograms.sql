CREATE TABLE t (i INT, s STRING, INDEX (i), INDEX(s))
WITH (sql_stats_automatic_collection_enabled = false);

-- Insert values 1-10,000.
INSERT INTO t SELECT i, i::STRING FROM generate_series(1, 10000) AS g(i);

-- Add 100 heavy hitters—every 100th value.
INSERT INTO t
SELECT i, i::STRING FROM (
  SELECT ((i//100)*100)+39 FROM generate_series(1, 10000) AS g(i)
) AS g(i);

-- Insert values 10,001-99,999 once.
INSERT INTO t SELECT i, i::STRING FROM generate_series(10001, 99999) AS g(i);

-- Insert values 100,00-110,000 with double the density.
INSERT INTO t SELECT i, i::STRING FROM generate_series(100000, 110000) AS g(i);
INSERT INTO t SELECT i, i::STRING FROM generate_series(100000, 110000) AS g(i);

-- Add 100 heavy hitters—every 100th-ish value.
INSERT INTO t
SELECT i, i::STRING FROM (
  SELECT ((i//100)*100)+17 FROM generate_series(100000, 110000) AS g(i)
) AS g(i);

CREATE STATISTICS stat0 ON i FROM t;
CREATE STATISTICS stat1 ON s FROM t;

SELECT statistics FROM [SHOW STATISTICS USING JSON FOR TABLE t];
