CREATE TABLE t (i INT, s STRING, INDEX (i), INDEX (s))
WITH (sql_stats_automatic_collection_enabled = false);

-- Insert values 1-10,000.
INSERT INTO t SELECT i, i::STRING FROM generate_series(1, 10000) AS g(i);

-- Add 100 heavy hitters—every 100th value between 1-10,000.
INSERT INTO t
SELECT i, i::STRING FROM (
  SELECT ((i//100)*100)+17 FROM generate_series(1, 10000) AS g(i)
) AS g(i);

-- Insert values 100,000-110,000.
INSERT INTO t SELECT i, i::STRING FROM generate_series(100000, 110000) AS g(i);

-- Add a heavy hitter at 105050.
INSERT INTO t SELECT 105050, 105050::STRING FROM generate_series(1, 1000);

-- Add a heavy hitter at 105099.
INSERT INTO t SELECT 105099, 105099::STRING FROM generate_series(1, 1000);

-- Delete values between 15050 and 15099 to simulate them not being randomly
-- sampled when collecting statistics.
DELETE FROM t WHERE i > 105050 AND i < 105099;

CREATE STATISTICS stat0 ON i FROM t;
CREATE STATISTICS stat1 ON s FROM t;

SELECT statistics FROM [SHOW STATISTICS USING JSON FOR TABLE t];
