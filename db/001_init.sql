-- 初始化脚本入口 - 由docker-entrypoint-initdb.d自动执行
\i /docker-entrypoint-initdb.d/init.sql
\i /docker-entrypoint-initdb.d/retention_policy.sql
