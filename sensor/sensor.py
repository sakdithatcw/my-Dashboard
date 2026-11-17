"""
อ่านค่าอุณหภูมิ/ความชื้นจาก DHT22 แล้วส่งไปที่ backend ทุก INTERVAL วินาที

ตั้งค่าผ่าน environment variable (ไม่ต้องแก้โค้ด):
  DEVICE_ID ชื่ออุปกรณ์ (ห้ามซ้ำกัน) (ค่าเริ่มต้น pi-01)
  API_URL   ที่อยู่ backend      (ค่าเริ่มต้น http://localhost:8000/api/readings)
  INTERVAL  ส่งทุกกี่วินาที        (ค่าเริ่มต้น 10)
  GPIO_PIN  ขาที่ต่อสาย DATA      (ค่าเริ่มต้น D4 = GPIO4)
  SIMULATE  ตั้งเป็น 1 เพื่อสุ่มค่าปลอม (ใช้ทดสอบบนคอมได้ ไม่ต้องมี sensor)
"""
import json
import os
import random
import time
import urllib.request

DEVICE_ID = os.getenv("DEVICE_ID", "pi-01")
API_URL = os.getenv("API_URL", "http://localhost:8000/api/readings")
INTERVAL = int(os.getenv("INTERVAL", "10"))
GPIO_PIN = os.getenv("GPIO_PIN", "D4")
SIMULATE = os.getenv("SIMULATE", "0") == "1"

dht = None

if SIMULATE:
    print("โหมดจำลอง: สุ่มค่าปลอม (ไม่ใช้ sensor จริง)")

    def read_sensor():
        return round(random.uniform(25, 32), 1), round(random.uniform(55, 80), 1)

else:
    import board
    import adafruit_dht

    dht = adafruit_dht.DHT22(getattr(board, GPIO_PIN))

    def read_sensor():
        return dht.temperature, dht.humidity


def send(temperature, humidity):
    body = json.dumps(
        {"device_id": DEVICE_ID, "temperature": temperature, "humidity": humidity}
    ).encode()
    req = urllib.request.Request(
        API_URL,
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=5) as res:
        return res.status


def main():
    print(f"อุปกรณ์ {DEVICE_ID} ส่งข้อมูลไปที่ {API_URL} ทุก {INTERVAL} วินาที (กด Ctrl+C เพื่อหยุด)")
    while True:
        try:
            t, h = read_sensor()
            if t is None or h is None:
                raise RuntimeError("sensor ไม่ส่งค่ากลับมา")
            status = send(t, h)
            print(f"ส่งแล้ว: {t:.1f} °C, {h:.1f} % (HTTP {status})")
        except RuntimeError as e:
            # DHT22 อ่านพลาดเป็นครั้งคราวเป็นเรื่องปกติ ข้ามไปรอบหน้าได้
            print("อ่านค่า sensor ไม่สำเร็จ (ปกติของ DHT):", e)
        except OSError as e:
            # เชื่อมต่อ backend ไม่ได้ / เน็ตหลุด
            print("ส่งข้อมูลไม่สำเร็จ:", e)
        time.sleep(INTERVAL)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nหยุดการทำงาน")
    finally:
        if dht is not None:
            dht.exit()
