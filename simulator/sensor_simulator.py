#!/usr/bin/env python3
import json
import time
import random
import argparse
import os
from datetime import datetime, timezone
from typing import Dict, Any

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("Error: paho-mqtt is required. Install with: pip install paho-mqtt")
    exit(1)

SOIL_TYPES = {
    "gravel": {"name": "砾石", "permeability": 1.0, "seepage_factor": 0.02},
    "sandy_loam": {"name": "砂壤土", "permeability": 0.4, "seepage_factor": 0.05},
    "clay": {"name": "粘土", "permeability": 0.02, "seepage_factor": 0.005},
    "loess": {"name": "黄土", "permeability": 0.15, "seepage_factor": 0.015},
    "sand": {"name": "砂土", "permeability": 0.8, "seepage_factor": 0.03},
    "silt": {"name": "粉土", "permeability": 0.1, "seepage_factor": 0.01},
    "rock": {"name": "岩石", "permeability": 0.001, "seepage_factor": 0.001},
}

CLIMATE_SCENARIOS = {
    "normal": {
        "name": "正常气候",
        "temp_range": (15, 35),
        "evaporation_range": (0.002, 0.008),
        "flow_base": 1.5,
        "flow_variance": 0.3,
        "water_level_base": 1.8,
        "water_level_variance": 0.2,
        "rain_factor": 1.0,
    },
    "drought": {
        "name": "极端干旱",
        "temp_range": (25, 45),
        "evaporation_range": (0.008, 0.02),
        "flow_base": 0.3,
        "flow_variance": 0.1,
        "water_level_base": 0.5,
        "water_level_variance": 0.1,
        "rain_factor": 0.1,
    },
    "flood": {
        "name": "暴雨洪水",
        "temp_range": (10, 25),
        "evaporation_range": (0.001, 0.003),
        "flow_base": 4.0,
        "flow_variance": 1.5,
        "water_level_base": 3.5,
        "water_level_variance": 0.8,
        "rain_factor": 5.0,
    },
    "freeze": {
        "name": "冬季冰冻",
        "temp_range": (-15, 5),
        "evaporation_range": (0.0005, 0.001),
        "flow_base": 0.8,
        "flow_variance": 0.2,
        "water_level_base": 2.0,
        "water_level_variance": 0.15,
        "rain_factor": 0.3,
    },
}


class KarezSensorSimulator:
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.mqtt_client = None
        self.karez_id = config.get("karez_id", 1)
        self.segment_count = config.get("segment_count", 10)
        self.shaft_count = config.get("shaft_count", 20)
        self.interval = config.get("interval", 5)
        self.soil_type = config.get("soil_type", "sandy_loam")
        self.climate_scenario = config.get("climate_scenario", "normal")
        self.anomaly_prob = config.get("anomaly_prob", 0.05)
        self.sedimentation_level = config.get("sedimentation_level", 0.0)

        if self.soil_type not in SOIL_TYPES:
            raise ValueError(f"Unknown soil type: {self.soil_type}. Available: {list(SOIL_TYPES.keys())}")
        if self.climate_scenario not in CLIMATE_SCENARIOS:
            raise ValueError(f"Unknown climate: {self.climate_scenario}. Available: {list(CLIMATE_SCENARIOS.keys())}")

        self.soil = SOIL_TYPES[self.soil_type]
        self.climate = CLIMATE_SCENARIOS[self.climate_scenario]
        self.message_count = 0

    def connect_mqtt(self):
        broker = self.config.get("mqtt_broker", "localhost")
        port = self.config.get("mqtt_port", 1883)
        client_id = f"karez-sim-{self.karez_id}-{int(time.time())}"

        self.mqtt_client = mqtt.Client(client_id=client_id)
        if self.config.get("mqtt_username"):
            self.mqtt_client.username_pw_set(
                self.config["mqtt_username"],
                self.config.get("mqtt_password")
            )

        self.mqtt_client.on_connect = self._on_connect
        self.mqtt_client.connect(broker, port, 60)
        self.mqtt_client.loop_start()
        time.sleep(1)

    def _on_connect(self, client, userdata, flags, rc):
        if rc == 0:
            print(f"Connected to MQTT broker at {self.config.get('mqtt_broker')}:{self.config.get('mqtt_port')}")
        else:
            print(f"MQTT connection failed with code {rc}")

    def disconnect(self):
        if self.mqtt_client:
            self.mqtt_client.loop_stop()
            self.mqtt_client.disconnect()

    def _generate_value(self, base: float, variance: float) -> float:
        value = base + random.uniform(-variance, variance)
        if random.random() < self.anomaly_prob:
            anomaly_factor = random.choice([0.2, 0.3, 1.8, 2.0, 2.5])
            value *= anomaly_factor
        return max(0, value)

    def _generate_temperature(self) -> float:
        t_min, t_max = self.climate["temp_range"]
        temp = random.uniform(t_min, t_max)
        return round(temp, 1)

    def _generate_evaporation(self, temperature: float) -> float:
        e_min, e_max = self.climate["evaporation_range"]
        base_evap = random.uniform(e_min, e_max)
        temp_factor = 1.0 + (temperature - 20) * 0.02
        soil_factor = 1.0 - self.soil["seepage_factor"]
        return round(base_evap * temp_factor * soil_factor * self.climate["rain_factor"], 6)

    def _generate_sensor_data(self, sensor_type: str, segment_id: int = None, shaft_id: int = None) -> Dict[str, Any]:
        now = datetime.now(timezone.utc)
        temp = self._generate_temperature()
        evap = self._generate_evaporation(temp)

        data = {
            "time": now.isoformat(),
            "karez_id": self.karez_id,
            "sensor_type": sensor_type,
            "sensor_id": f"{sensor_type}_{segment_id or shaft_id or 1:03d}",
            "temperature": temp,
            "evaporation": evap,
            "soil_type": self.soil_type,
            "climate_scenario": self.climate_scenario,
        }

        if segment_id is not None:
            data["segment_id"] = segment_id
            sedimentation_effect = 1.0 - self.sedimentation_level * 0.7

            if sensor_type == "flow":
                flow = self._generate_value(
                    self.climate["flow_base"] * sedimentation_effect,
                    self.climate["flow_variance"]
                )
                data["flow_rate"] = round(flow, 4)
                data["velocity"] = round(flow / 2.5 + random.uniform(-0.1, 0.1), 3)
                data["turbidity"] = round(random.uniform(5, 50) + self.sedimentation_level * 100, 1)

            elif sensor_type == "water_level":
                wl = self._generate_value(
                    self.climate["water_level_base"],
                    self.climate["water_level_variance"]
                )
                data["water_level"] = round(wl, 3)

        if shaft_id is not None:
            data["shaft_id"] = shaft_id
            if sensor_type == "shaft_water_level":
                base_wl = self.climate["water_level_base"] * 0.9
                wl = self._generate_value(base_wl, 0.15)
                data["shaft_water_level"] = round(wl, 3)

        return data

    def publish_data(self, topic: str, data: Dict[str, Any]):
        if self.mqtt_client:
            payload = json.dumps(data, ensure_ascii=False)
            self.mqtt_client.publish(topic, payload, qos=1)
            self.message_count += 1
            if self.message_count % 100 == 0:
                print(f"Published {self.message_count} messages...")

    def run_once(self):
        topic_prefix = self.config.get("mqtt_topic_prefix", "karez")

        for seg_id in range(1, self.segment_count + 1):
            flow_data = self._generate_sensor_data("flow", segment_id=seg_id)
            self.publish_data(f"{topic_prefix}/sensor/flow/{seg_id}", flow_data)

            wl_data = self._generate_sensor_data("water_level", segment_id=seg_id)
            self.publish_data(f"{topic_prefix}/sensor/water_level/{seg_id}", wl_data)

        for shaft_id in range(1, self.shaft_count + 1):
            swl_data = self._generate_sensor_data("shaft_water_level", shaft_id=shaft_id)
            self.publish_data(f"{topic_prefix}/sensor/shaft_water_level/{shaft_id}", swl_data)

    def run(self, duration: int = None):
        print(f"\n{'='*60}")
        print(f"坎儿井传感器模拟器启动")
        print(f"{'='*60}")
        print(f"坎儿井ID: {self.karez_id}")
        print(f"土壤类型: {self.soil['name']} ({self.soil_type})")
        print(f"  渗透系数: {self.soil['permeability']}")
        print(f"  渗流因子: {self.soil['seepage_factor']}")
        print(f"气候场景: {self.climate['name']} ({self.climate_scenario})")
        print(f"  温度范围: {self.climate['temp_range']}℃")
        print(f"  基础流量: {self.climate['flow_base']} m³/s")
        print(f"  降雨系数: {self.climate['rain_factor']}")
        print(f"淤塞程度: {self.sedimentation_level * 100:.0f}%")
        print(f"异常概率: {self.anomaly_prob * 100:.0f}%")
        print(f"发送间隔: {self.interval}s")
        print(f"暗渠段数: {self.segment_count}")
        print(f"竖井数量: {self.shaft_count}")
        print(f"{'='*60}\n")

        start_time = time.time()
        try:
            while True:
                self.run_once()
                time.sleep(self.interval)

                if duration and (time.time() - start_time) >= duration:
                    print(f"\n达到运行时长 {duration}s，停止模拟")
                    break
        except KeyboardInterrupt:
            print(f"\n用户中断，共发送 {self.message_count} 条消息")
        finally:
            self.disconnect()


def main():
    parser = argparse.ArgumentParser(
        description="坎儿井传感器模拟器 - 支持多种土壤类型和气候场景",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=f"""
土壤类型: {', '.join(SOIL_TYPES.keys())}
气候场景: {', '.join(CLIMATE_SCENARIOS.keys())}

示例:
  # 正常气候 + 砂壤土
  python sensor_simulator.py --soil-type sandy_loam --climate normal

  # 极端干旱 + 粘土
  python sensor_simulator.py --soil-type clay --climate drought --interval 3

  # 暴雨洪水 + 砾石 + 30%淤塞
  python sensor_simulator.py --soil-type gravel --climate flood --sedimentation 0.3

  # 运行60秒后自动停止
  python sensor_simulator.py --duration 60
        """
    )

    parser.add_argument("--karez-id", type=int, default=1, help="坎儿井ID")
    parser.add_argument("--segment-count", type=int, default=10, help="暗渠段数")
    parser.add_argument("--shaft-count", type=int, default=20, help="竖井数量")
    parser.add_argument("--interval", type=int, default=5, help="发送间隔(秒)")
    parser.add_argument("--soil-type", type=str, default="sandy_loam",
                        help=f"土壤类型: {', '.join(SOIL_TYPES.keys())}")
    parser.add_argument("--climate-scenario", type=str, default="normal",
                        help=f"气候场景: {', '.join(CLIMATE_SCENARIOS.keys())}")
    parser.add_argument("--anomaly-prob", type=float, default=0.05,
                        help="数据异常概率 (0-1)")
    parser.add_argument("--sedimentation", type=float, default=0.0,
                        help="淤塞程度 (0-1)")
    parser.add_argument("--mqtt-broker", type=str, default=os.environ.get("MQTT_BROKER", "localhost"),
                        help="MQTT broker地址")
    parser.add_argument("--mqtt-port", type=int, default=int(os.environ.get("MQTT_PORT", "1883")),
                        help="MQTT broker端口")
    parser.add_argument("--mqtt-username", type=str, default=os.environ.get("MQTT_USERNAME"),
                        help="MQTT用户名")
    parser.add_argument("--mqtt-password", type=str, default=os.environ.get("MQTT_PASSWORD"),
                        help="MQTT密码")
    parser.add_argument("--mqtt-topic-prefix", type=str, default="karez",
                        help="MQTT主题前缀")
    parser.add_argument("--duration", type=int, default=None,
                        help="运行时长(秒)，None表示一直运行")

    args = parser.parse_args()

    config = vars(args)
    config.pop("func", None)

    try:
        sim = KarezSensorSimulator(config)
        sim.connect_mqtt()
        sim.run(args.duration)
    except ValueError as e:
        print(f"配置错误: {e}")
        exit(1)
    except Exception as e:
        print(f"错误: {e}")
        exit(1)


if __name__ == "__main__":
    main()
