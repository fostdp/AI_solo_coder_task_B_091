-- ============================================
-- TimescaleDB 降采样和保留策略配置
-- ============================================

-- ============================================
-- 1. 数据保留策略 (Retention Policies)
--    原始数据保留30天，1小时聚合保留1年，1天聚合永久保留
-- ============================================

-- 传感器原始数据：保留30天
SELECT add_retention_policy('sensor_data', INTERVAL '30 days', if_not_exists => TRUE);

-- 仿真结果原始数据：保留90天
SELECT add_retention_policy('simulation_results', INTERVAL '90 days', if_not_exists => TRUE);

-- 分配结果原始数据：保留1年
SELECT add_retention_policy('allocation_results', INTERVAL '1 year', if_not_exists => TRUE);

-- 告警事件：保留2年
SELECT add_retention_policy('alert_events', INTERVAL '2 years', if_not_exists => TRUE);

-- ============================================
-- 2. 连续聚合视图 (Continuous Aggregates)
--    自动后台聚合，支持快速查询历史趋势
-- ============================================

-- 每小时传感器数据聚合
CREATE MATERIALIZED VIEW IF NOT EXISTS sensor_data_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    karez_id,
    segment_id,
    sensor_type,
    COUNT(*) AS sample_count,
    AVG(flow_rate) AS avg_flow_rate,
    MAX(flow_rate) AS max_flow_rate,
    MIN(flow_rate) AS min_flow_rate,
    AVG(water_level) AS avg_water_level,
    MAX(water_level) AS max_water_level,
    MIN(water_level) AS min_water_level,
    AVG(shaft_water_level) AS avg_shaft_water_level,
    AVG(evaporation) AS avg_evaporation,
    SUM(evaporation) AS total_evaporation,
    AVG(temperature) AS avg_temperature,
    MAX(temperature) AS max_temperature,
    MIN(temperature) AS min_temperature,
    AVG(turbidity) AS avg_turbidity,
    AVG(velocity) AS avg_velocity
FROM sensor_data
GROUP BY bucket, karez_id, segment_id, sensor_type
WITH NO DATA;

-- 每天传感器数据聚合
CREATE MATERIALIZED VIEW IF NOT EXISTS sensor_data_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    karez_id,
    segment_id,
    sensor_type,
    COUNT(*) AS sample_count,
    AVG(flow_rate) AS avg_flow_rate,
    MAX(flow_rate) AS max_flow_rate,
    MIN(flow_rate) AS min_flow_rate,
    AVG(water_level) AS avg_water_level,
    MAX(water_level) AS max_water_level,
    MIN(water_level) AS min_water_level,
    AVG(shaft_water_level) AS avg_shaft_water_level,
    AVG(evaporation) AS avg_evaporation,
    SUM(evaporation) AS total_evaporation,
    AVG(temperature) AS avg_temperature,
    MAX(temperature) AS max_temperature,
    MIN(temperature) AS min_temperature,
    AVG(turbidity) AS avg_turbidity,
    AVG(velocity) AS avg_velocity
FROM sensor_data
GROUP BY bucket, karez_id, segment_id, sensor_type
WITH NO DATA;

-- 每小时水力损失聚合
CREATE MATERIALIZED VIEW IF NOT EXISTS simulation_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    karez_id,
    segment_id,
    AVG(inflow_rate) AS avg_inflow_rate,
    AVG(outflow_rate) AS avg_outflow_rate,
    SUM(seepage_loss) AS total_seepage_loss,
    SUM(evaporation_loss) AS total_evaporation_loss,
    SUM(total_loss) AS total_loss,
    AVG(water_depth) AS avg_water_depth,
    AVG(reynolds_number) AS avg_reynolds_number,
    AVG(froude_number) AS avg_froude_number
FROM simulation_results
GROUP BY bucket, karez_id, segment_id
WITH NO DATA;

-- 每天水力损失聚合
CREATE MATERIALIZED VIEW IF NOT EXISTS simulation_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    karez_id,
    segment_id,
    AVG(inflow_rate) AS avg_inflow_rate,
    AVG(outflow_rate) AS avg_outflow_rate,
    SUM(seepage_loss) AS total_seepage_loss,
    SUM(evaporation_loss) AS total_evaporation_loss,
    SUM(total_loss) AS total_loss,
    AVG(water_depth) AS avg_water_depth,
    AVG(reynolds_number) AS avg_reynolds_number,
    AVG(froude_number) AS avg_froude_number
FROM simulation_results
GROUP BY bucket, karez_id, segment_id
WITH NO DATA;

-- 每天水量分配聚合
CREATE MATERIALIZED VIEW IF NOT EXISTS allocation_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    karez_id,
    oasis_id,
    branch_channel_id,
    AVG(allocated_flow) AS avg_allocated_flow,
    AVG(allocation_ratio) AS avg_allocation_ratio,
    AVG(demand_met) AS avg_demand_met,
    AVG(objective_value) AS avg_objective_value
FROM allocation_results
GROUP BY bucket, karez_id, oasis_id, branch_channel_id
WITH NO DATA;

-- ============================================
-- 3. 聚合刷新策略 (Refresh Policies)
--    定时刷新连续聚合视图
-- ============================================

-- 小时级聚合：每30分钟刷新一次，覆盖最近2小时数据
SELECT add_continuous_aggregate_policy('sensor_data_hourly',
    start_offset => INTERVAL '2 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '30 minutes',
    if_not_exists => TRUE);

-- 日级聚合：每天刷新一次，覆盖最近3天数据
SELECT add_continuous_aggregate_policy('sensor_data_daily',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE);

-- 仿真小时级聚合：每小时刷新
SELECT add_continuous_aggregate_policy('simulation_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

-- 仿真日级聚合：每天刷新
SELECT add_continuous_aggregate_policy('simulation_daily',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE);

-- 分配日级聚合：每天刷新
SELECT add_continuous_aggregate_policy('allocation_daily',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE);

-- ============================================
-- 4. 聚合视图保留策略
-- ============================================

-- 小时级聚合：保留1年
SELECT add_retention_policy('sensor_data_hourly', INTERVAL '1 year', if_not_exists => TRUE);
SELECT add_retention_policy('simulation_hourly', INTERVAL '1 year', if_not_exists => TRUE);

-- 日级聚合：永久保留（不设置保留策略）

-- ============================================
-- 5. 压缩策略 (Compression)
--    对历史数据启用列式压缩，节省存储空间
-- ============================================

-- 启用sensor_data表压缩
ALTER TABLE sensor_data SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'karez_id, segment_id, sensor_type',
    timescaledb.compress_orderby = 'time DESC'
);

-- 启用simulation_results表压缩
ALTER TABLE simulation_results SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'karez_id, segment_id',
    timescaledb.compress_orderby = 'time DESC'
);

-- 启用allocation_results表压缩
ALTER TABLE allocation_results SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'karez_id, oasis_id',
    timescaledb.compress_orderby = 'time DESC'
);

-- 启用alert_events表压缩
ALTER TABLE alert_events SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'karez_id, alert_type, alert_level',
    timescaledb.compress_orderby = 'time DESC'
);

-- 添加压缩策略：超过7天的数据自动压缩
SELECT add_compression_policy('sensor_data', INTERVAL '7 days', if_not_exists => TRUE);
SELECT add_compression_policy('simulation_results', INTERVAL '14 days', if_not_exists => TRUE);
SELECT add_compression_policy('allocation_results', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_compression_policy('alert_events', INTERVAL '90 days', if_not_exists => TRUE);

-- ============================================
-- 6. 重新排序策略 (Reorder)
--    按时间排序优化查询性能
-- ============================================

SELECT add_reorder_policy('sensor_data', 'idx_sensor_data_karez', if_not_exists => TRUE);
SELECT add_reorder_policy('simulation_results', 'idx_sim_results_karez', if_not_exists => TRUE);

-- ============================================
-- 查看已配置的策略
-- ============================================

-- SELECT * FROM timescaledb_information.jobs;
-- SELECT * FROM timescaledb_information.continuous_aggregates;
-- SELECT * FROM timescaledb_information.retention_policies;
-- SELECT * FROM timescaledb_information.compression_settings;
