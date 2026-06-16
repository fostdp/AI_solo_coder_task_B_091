# 古代坎儿井暗渠水力仿真与水量分配系统

一套完整的坎儿井水文仿真与水资源优化分配全栈系统。基于明渠均匀流理论、渗流力学和线性规划算法，实现对古代坎儿井暗渠输水过程的高精度仿真与下游绿洲水量的智能优化分配。

---

## 目录

- [系统架构](#系统架构)
- [技术栈](#技术栈)
- [核心功能](#核心功能)
- [快速开始](#快速开始)
- [部署指南](#部署指南)
- [传感器模拟器](#传感器模拟器)
- [监控与运维](#监控与运维)
- [API 文档](#api-文档)
- [配置说明](#配置说明)
- [目录结构](#目录结构)

---

## 系统架构

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                      前端层                                          │
│  [Nginx + Gzip/Brotli]                                                              │
│  ┌────────────────────┐  ┌───────────────────────────────────┐                        │
│  │  karez3d.js        │  │  allocation_panel.js             │                        │
│  │  (Three.js 3D渲染) │  │  (水量分配面板)                  │                        │
│  └─────────┬──────────┘  └───────────────┬───────────────────┘                        │
│            │                              │                                                │
│            └──────────────┬───────────────┘                                                │
│                           │                                                                │
└───────────────────────────┼────────────────────────────────────────────────────────────────┘
                            │ REST API (8080)
┌───────────────────────────┼────────────────────────────────────────────────────────────────┐
│                         后端层 (Go Actor 模型)                                              │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  │
│  │  dtu_receiver    │──│ hydraulic_sim    │──│ water_allocator  │──│ alarm_mqtt       │  │
│  │  (数据采集校验)  │  │  (水力计算渗流)  │  │  (线性规划优化)  │  │  (告警评估推送)  │  │
│  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘  │
│           │                     │                     │                     │            │
│           └─────────────────────┴─────────┬───────────┴─────────────────────┘            │
│                                           │  Channel 通信                               │
│  ┌────────────────────────────────────────┴──────────────────────────────────────┐      │
│  │  Gin Web 框架  │  Prometheus 指标  │  pprof 性能分析  │  健康检查            │      │
│  └──────────────────────────────────────────────────────────────────────────────┘      │
└───────────────┬───────────────────┬──────────────────────────┬───────────────────────┘
                │                   │                          │
                │ MQTT              │ JDBC                     │ HTTP 指标
┌───────────────┴──────┐  ┌────────┴──────────────┐  ┌────────┴───────────────┐
│  MQTT Broker        │  │  TimescaleDB          │  │  Prometheus + Grafana   │
│  (Eclipse Mosquitto)│  │  (时序数据库)         │  │  (监控可视化)           │
│  Port: 1883/9001    │  │  Port: 5432           │  │  Port: 9090/3000        │
└──────────────────────┘  └───────────────────────┘  └─────────────────────────┘
                ▲
                │ MQTT 传感器数据
┌───────────────┴──────────────────────────────────────────────────────────────┐
│  传感器模拟器 (Python)                                                        │
│  支持 7 种土壤类型 + 4 种气候场景 + 淤塞模拟 + 异常数据注入                    │
└───────────────────────────────────────────────────────────────────────────────┘
```

### Go 模块拆分 (Actor 模型)

```
sensorDataChan (100)
     │
     ▼
┌──────────────┐     simRequestChan (10)     ┌──────────────┐
│ dtu_receiver │ ──────────────────────────> │ hydraulic_sim│
└──────────────┘                              └──────┬───────┘
     │                                               │ simResultChan (20)
     │ MQTT Subscribe                                │
     ▼                                               ▼
┌──────────────┐     allocRequestChan (10)     ┌──────────────┐
│   handlers   │ ──────────────────────────> │water_allocator│
└──────────────┘                              └──────┬───────┘
     │                                               │ allocResultChan (10)
     │                                               │
     │ alarmRequestChan (10)                         ▼
     └─────────────────────────────────────>  ┌──────────────┐
                                              │  alarm_mqtt  │
                                              └──────┬───────┘
                                                     │ MQTT Publish
                                                     ▼
                                                  告警推送
```

---

## 技术栈

### 后端
- **Go 1.22**: 主开发语言，静态编译
- **Gin**: Web 框架
- **TimescaleDB (PostgreSQL 16)**: 时序数据库
- **Eclipse Mosquitto**: MQTT 消息代理
- **pgx/v5**: PostgreSQL 驱动
- **Prometheus Client**: 指标采集
- **net/http/pprof**: 性能分析

### 前端
- **Three.js**: 3D 可视化
- **Canvas**: 粒子动画
- **Nginx**: 静态资源服务（Gzip/Brotli 压缩）

### 监控
- **Prometheus 2.52**: 指标存储与告警
- **Grafana 10.4**: 可视化面板

### 部署
- **Docker**: 容器化
- **Docker Compose**: 服务编排
- **Alpine**: 基础镜像（最小化体积）

---

## 核心功能

### 1. 暗渠水力仿真
- **明渠均匀流计算**: 基于曼宁公式
- **渗流分析**: 7种土壤类型渗透系数修正
- **水头损失计算**: 沿程损失+局部损失
- **流态判定**: 雷诺数+弗劳德数
- **淤塞风险评估**: 基于流速和浊度

### 2. 水量分配优化
- **线性规划**: 自定义单纯形法求解器
- **Big-M 松弛变量**: 极端干旱时保证有可行解
- **多目标优化**: 公平性+效率+优先级
- **优先级加权**: 不同绿洲差异化权重

### 3. 告警系统
- **淤塞告警**: 流速过低触发
- **水量不足告警**: 水位/流量过低触发
- **MQTT 推送**: 实时告警通知
- **冷却机制**: 30分钟去重

### 4. 前端可视化
- **3D 暗渠剖面**: Three.js 透明管道渲染
- **LOD 三级细节**: 近/中/远景自适应
- **水流粒子动画**: 实时流量可视化
- **水量分配面板**: 各绿洲满足率进度条

### 5. 工程化特性
- **多阶段构建**: Go 静态编译镜像 < 20MB
- **Channel 通信**: Actor 并发模型
- **配置外置**: JSON 参数化配置
- **健康检查**: Docker 自动健康检测
- **数据压缩**: TimescaleDB 列式压缩
- **降采样**: 小时级/日级连续聚合
- **保留策略**: 原始数据30天，聚合数据1年

---

## 快速开始

### 前置要求
- Docker >= 24.0
- Docker Compose >= 2.20
- 至少 4GB 可用内存
- 至少 10GB 可用磁盘空间

### 一键启动

```bash
# 1. 克隆项目
git clone <repository-url>
cd AI_solo_coder_task_A_091

# 2. 启动核心服务（不含模拟器）
docker-compose up -d

# 3. 等待所有服务健康
docker-compose ps

# 4. 启动传感器模拟器（可选，按需选择土壤和气候）
docker-compose --profile simulator up -d sensor-simulator

# 5. 访问应用
# 前端: http://localhost:8088
# 后端API: http://localhost:8080
# Grafana: http://localhost:3000 (admin/admin123)
# Prometheus: http://localhost:9090
# Pprof: http://localhost:6060/debug/pprof/
```

### 验证服务

```bash
# 检查健康状态
curl http://localhost:8080/health
# {"status":"ok","timestamp":"...","version":"1.0.0"}

# 检查Prometheus指标
curl http://localhost:8080/metrics | head -20

# 获取仪表盘数据
curl http://localhost:8080/api/karez/1/dashboard
```

---

## 部署指南

### 环境变量配置

复制 `.env` 并根据需要修改：

```bash
cp .env .env.local
# 编辑 .env.local 修改密码等配置
```

主要环境变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `POSTGRES_DB` | karez | 数据库名 |
| `POSTGRES_USER` | karez | 数据库用户 |
| `POSTGRES_PASSWORD` | karez123 | 数据库密码 |
| `MQTT_BROKER` | mosquitto | MQTT Broker地址 |
| `MQTT_PORT` | 1883 | MQTT端口 |
| `SERVER_PORT` | 8080 | 后端服务端口 |
| `SIMULATION_INTERVAL` | 300 | 仿真间隔(秒) |
| `ALERT_CHECK_INTERVAL` | 60 | 告警检查间隔(秒) |
| `TZ` | Asia/Shanghai | 时区 |

### 生产部署建议

1. **使用 Traefik/Nginx 作为反向代理**
2. **启用 HTTPS**（Let's Encrypt）
3. **配置防火墙**，仅开放必要端口
4. **修改默认密码**（数据库、Grafana等）
5. **配置数据备份**（TimescaleDB 定时备份）
6. **启用 Mosquitto 认证**
7. **配置监控告警通知**（邮件/钉钉/企业微信）

### 常用运维命令

```bash
# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f backend
docker-compose logs -f timescaledb
docker-compose logs -f mosquitto

# 重启服务
docker-compose restart backend

# 更新服务
docker-compose pull
docker-compose up -d

# 停止所有服务
docker-compose down

# 停止服务并清除数据（谨慎使用）
docker-compose down -v

# 进入容器
docker-compose exec timescaledb psql -U karez -d karez
```

---

## 传感器模拟器

### 功能特性

- **7 种土壤类型**: 砾石、砂壤土、粘土、黄土、砂土、粉土、岩石
- **4 种气候场景**: 正常气候、极端干旱、暴雨洪水、冬季冰冻
- **淤塞模拟**: 0-100% 淤塞程度，影响过流能力
- **异常数据注入**: 可配置概率生成异常值
- **实时参数**: 流量、水位、竖井水位、温度、蒸发量、浊度、流速

### 命令行参数

```bash
python simulator/sensor_simulator.py \
  --soil-type <soil_type> \
  --climate-scenario <scenario> \
  --sedimentation <0-1> \
  --anomaly-prob <0-1> \
  --interval <seconds> \
  --duration <seconds>
```

### 土壤类型参数

| 土壤类型 | 渗透系数 | 渗流因子 | 说明 |
|---------|----------|----------|------|
| `gravel` | 1.0 | 0.02 | 砾石，渗透性强 |
| `sandy_loam` | 0.4 | 0.05 | 砂壤土，中等渗透 |
| `clay` | 0.02 | 0.005 | 粘土，低渗透 |
| `loess` | 0.15 | 0.015 | 黄土，中等偏低 |
| `sand` | 0.8 | 0.03 | 砂土，渗透性较强 |
| `silt` | 0.1 | 0.01 | 粉土，渗透性较弱 |
| `rock` | 0.001 | 0.001 | 岩石，几乎不渗透 |

### 气候场景参数

| 场景 | 温度范围 | 基础流量 | 降雨系数 | 说明 |
|------|----------|----------|----------|------|
| `normal` | 15-35℃ | 1.5 m³/s | 1.0 | 正常气候 |
| `drought` | 25-45℃ | 0.3 m³/s | 0.1 | 极端干旱 |
| `flood` | 10-25℃ | 4.0 m³/s | 5.0 | 暴雨洪水 |
| `freeze` | -15-5℃ | 0.8 m³/s | 0.3 | 冬季冰冻 |

### 使用示例

```bash
# 示例1: 正常气候 + 砂壤土
python simulator/sensor_simulator.py \
  --soil-type sandy_loam \
  --climate-scenario normal \
  --interval 5

# 示例2: 极端干旱 + 粘土（测试水资源分配）
python simulator/sensor_simulator.py \
  --soil-type clay \
  --climate-scenario drought \
  --interval 3 \
  --anomaly-prob 0.1

# 示例3: 暴雨洪水 + 砾石 + 30%淤塞（测试淤塞告警）
python simulator/sensor_simulator.py \
  --soil-type gravel \
  --climate-scenario flood \
  --sedimentation 0.3 \
  --interval 5

# 示例4: 冬季冰冻 + 黄土（测试低温场景）
python simulator/sensor_simulator.py \
  --soil-type loess \
  --climate-scenario freeze \
  --interval 10

# 示例5: Docker 方式运行
docker-compose run --rm sensor-simulator \
  --soil-type clay \
  --climate-scenario drought \
  --duration 3600
```

### MQTT 主题

模拟器发布的 MQTT 主题：
- `karez/sensor/flow/{segment_id}` - 流量传感器数据
- `karez/sensor/water_level/{segment_id}` - 水位传感器数据
- `karez/sensor/shaft_water_level/{shaft_id}` - 竖井水位数据

### 消息格式

```json
{
  "time": "2024-01-15T10:30:00+08:00",
  "karez_id": 1,
  "segment_id": 1,
  "sensor_type": "flow",
  "sensor_id": "flow_001",
  "flow_rate": 1.5234,
  "water_level": 1.85,
  "temperature": 25.5,
  "evaporation": 0.0045,
  "velocity": 0.609,
  "turbidity": 25.3,
  "soil_type": "sandy_loam",
  "climate_scenario": "normal"
}
```

---

## 监控与运维

### Prometheus 指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `karez_http_requests_total` | Counter | HTTP 请求总数 |
| `karez_http_request_duration_seconds` | Histogram | HTTP 请求延迟 |
| `karez_sensor_data_received_total` | Counter | 接收传感器数据总数 |
| `karez_sensor_data_valid_total` | Counter | 有效传感器数据总数 |
| `karez_sensor_data_invalid_total` | Counter | 无效传感器数据总数 |
| `karez_simulations_run_total` | Counter | 水力仿真运行次数 |
| `karez_allocations_run_total` | Counter | 水量分配运行次数 |
| `karez_alerts_triggered_total` | Counter | 告警触发总数 |
| `karez_mqtt_messages_published_total` | Counter | MQTT发布消息数 |
| `karez_mqtt_messages_received_total` | Counter | MQTT接收消息数 |
| `karez_channel_backpressure` | Gauge | Channel 拥塞度(0-1) |
| `karez_database_query_duration_seconds` | Histogram | 数据库查询延迟 |

### Pprof 性能分析

```bash
# 查看堆内存分配
go tool pprof http://localhost:6060/debug/pprof/heap

# 查看CPU使用
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 查看goroutine
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 查看内存分配
go tool pprof http://localhost:6060/debug/pprof/allocs

# 查看锁竞争
go tool pprof http://localhost:6060/debug/pprof/mutex
```

### Grafana 仪表盘

访问 http://localhost:3000 (admin/admin123)

已预置仪表盘：
- **坎儿井系统概览**: HTTP请求、延迟、传感器数据速率、告警频率

可自定义仪表盘：
- 各暗渠段流量趋势
- 水位变化曲线
- 渗流损失统计
- 水量分配结果
- 告警统计

### TimescaleDB 数据管理

```sql
-- 查看保留策略
SELECT * FROM timescaledb_information.retention_policies;

-- 查看压缩策略
SELECT * FROM timescaledb_information.compression_settings;

-- 查看连续聚合
SELECT * FROM timescaledb_information.continuous_aggregates;

-- 查看后台任务
SELECT * FROM timescaledb_information.jobs;

-- 查询小时级聚合数据
SELECT * FROM sensor_data_hourly
WHERE bucket >= NOW() - INTERVAL '7 days'
ORDER BY bucket DESC;

-- 查询日级损失统计
SELECT * FROM simulation_daily
WHERE bucket >= NOW() - INTERVAL '30 days'
ORDER BY bucket DESC;

-- 手动压缩chunk
SELECT compress_chunk(i, if_not_compressed => true)
FROM show_chunks('sensor_data', older_than => INTERVAL '7 days') i;

-- 手动删除过期数据
SELECT drop_chunks('sensor_data', older_than => INTERVAL '30 days');
```

---

## API 文档

### 健康检查
```
GET /health
```

### 指标采集
```
GET /metrics
```

### 基础信息查询
```
GET /api/karez
GET /api/karez/{karez_id}/segments
GET /api/karez/{karez_id}/shafts
GET /api/karez/{karez_id}/branches
GET /api/karez/{karez_id}/oases
GET /api/karez/{karez_id}/dashboard
```

### 传感器数据
```
POST /api/sensor
GET /api/sensor/{karez_id}/latest
GET /api/sensor/{karez_id}/range?start=...&end=...
```

### 水力仿真
```
POST /api/simulate
POST /api/simulate/hydraulic
```

### 水量分配
```
POST /api/allocate
```

### 告警管理
```
GET /api/alerts/{karez_id}
POST /api/alerts/check/{karez_id}
POST /api/alerts/acknowledge
POST /api/alerts/resolve
```

---

## 配置说明

### 水力参数配置

[backend/config/hydraulic_params.json](backend/config/hydraulic_params.json)

```json
{
  "default_channel": {
    "width": 0.8,
    "height": 1.2,
    "slope": 0.006,
    "roughness_coeff": 0.013
  },
  "soil_permeability": {
    "gravel": {"permeability": 1.0, "correction_factor": 1.8},
    "sandy_loam": {"permeability": 0.4, "correction_factor": 1.2},
    ...
  },
  "sedimentation_thresholds": {
    "low": 0.5,
    "medium": 0.3,
    "high": 0.1
  },
  "evaporation": {
    "base_rate": 0.005,
    "temp_coeff": 0.0002
  },
  "simulation": {
    "default_inflow_rate": 1.5,
    "time_step": 60,
    "max_iterations": 100
  }
}
```

### 农业需水配置

[backend/config/agriculture_demand.json](backend/config/agriculture_demand.json)

```json
{
  "crop_types": {
    "grape": {"name": "葡萄", "water_requirement_coeff": 1.2},
    "cotton": {"name": "棉花", "water_requirement_coeff": 0.9},
    "wheat": {"name": "小麦", "water_requirement_coeff": 0.7},
    ...
  },
  "oasis_defaults": {
    "min_allocation_ratio": 0.3,
    "demand_update_interval": 86400
  },
  "allocation_algorithm": {
    "big_m": 1000000,
    "min_allocation_ratio": 0.3,
    "max_iterations": 10000,
    "tolerance": 1e-9
  },
  "water_shortage_levels": {
    "mild": 0.8,
    "moderate": 0.5,
    "severe": 0.3
  }
}
```

---

## 目录结构

```
AI_solo_coder_task_A_091/
├── backend/                    # Go 后端服务
│   ├── alarm_mqtt/            # 告警MQTT模块
│   ├── config/                 # 配置加载
│   │   ├── config.go
│   │   ├── hydraulic_params.json    # 水力参数
│   │   └── agriculture_demand.json  # 农业需水参数
│   ├── db/                     # 数据库层
│   ├── dtu_receiver/           # 数据采集模块
│   ├── handlers/               # API处理器
│   ├── hydraulic_sim/          # 水力仿真模块
│   ├── metrics/                # Prometheus指标
│   ├── models/                 # 数据模型
│   ├── mqtt/                   # MQTT客户端
│   ├── water_allocator/        # 水量分配模块
│   ├── Dockerfile              # 多阶段构建
│   ├── main.go                 # 入口文件
│   └── go.mod
├── frontend/                   # 前端应用
│   ├── js/
│   │   ├── karez3d.js         # 3D可视化模块
│   │   ├── allocation_panel.js # 分配面板模块
│   │   └── main.js             # 主逻辑
│   ├── css/
│   ├── index.html
│   ├── nginx.conf              # Nginx配置(Gzip/Brotli)
│   └── Dockerfile
├── simulator/                  # 传感器模拟器
│   ├── sensor_simulator.py
│   ├── requirements.txt
│   └── Dockerfile
├── db/                         # 数据库脚本
│   ├── init.sql                # 初始化脚本
│   ├── retention_policy.sql    # 保留策略&降采样
│   └── 001_init.sql            # 入口脚本
├── mqtt/                       # MQTT配置
│   └── mosquitto.conf
├── monitoring/                 # 监控配置
│   ├── prometheus.yml
│   ├── rules/
│   │   └── alerts.yml          # 告警规则
│   └── grafana/
│       ├── provisioning/
│       │   ├── datasources/
│       │   └── dashboards/
│       └── dashboards/
│           └── karez-overview.json
├── docker-compose.yml          # 服务编排
├── .env                        # 环境变量
└── README.md
```

---

## 开发指南

### 本地开发

```bash
# 后端开发
cd backend
$env:GONOSUMDB='*'
go run .

# 前端开发
# 直接用浏览器打开 frontend/index.html，或使用本地HTTP服务器
python -m http.server 8080 --directory frontend

# 启动依赖服务（仅数据库和MQTT）
docker-compose up -d timescaledb mosquitto

# 运行模拟器
cd simulator
pip install -r requirements.txt
python sensor_simulator.py --soil-type sandy_loam --climate normal
```

### 添加新的土壤类型

1. 在 [backend/config/hydraulic_params.json](backend/config/hydraulic_params.json) 中添加参数
2. 在 [simulator/sensor_simulator.py](simulator/sensor_simulator.py) 的 `SOIL_TYPES` 中添加
3. 在 [db/init.sql](db/init.sql) 的 `soil_permeability` 表中添加

### 添加新的气候场景

1. 在 [simulator/sensor_simulator.py](simulator/sensor_simulator.py) 的 `CLIMATE_SCENARIOS` 中添加
2. 更新帮助文档

---

## License

MIT License

---

## 参考资料

- [曼宁公式](https://en.wikipedia.org/wiki/Manning%27s_equation)
- [渗流力学](https://en.wikipedia.org/wiki/Fluid_flow_in_porous_media)
- [线性规划](https://en.wikipedia.org/wiki/Linear_programming)
- [单纯形法](https://en.wikipedia.org/wiki/Simplex_algorithm)
- [TimescaleDB 文档](https://docs.timescale.com/)
- [Three.js 文档](https://threejs.org/docs/)
