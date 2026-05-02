-- ClickHouse visit_stats table
-- Populated by Flink job (DwsShortLinkVisitStatsApp)

CREATE DATABASE IF NOT EXISTS visit_stats;

CREATE TABLE IF NOT EXISTS visit_stats.visit_stats
(
    code          String,
    referer       String,
    is_new        String,
    account_no    UInt64,
    province      String,
    city          String,
    ip            String,
    browser_name  String,
    os            String,
    device_type   String,
    pv            UInt64,
    uv            UInt64,
    start_time    DateTime,
    end_time      DateTime,
    ts            UInt64
)
ENGINE = MergeTree()
ORDER BY (account_no, code, ts);
