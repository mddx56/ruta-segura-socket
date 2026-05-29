import socket
import struct
import binascii
import time
import math
import random
from datetime import datetime

# --- CONFIGURACIÓN ---
# HOST = "161.35.242.117"
HOST = "localhost"
PORT = 5050
IMEI_TARGET = "0864209042345678"

# Configuración de Simulación
UPDATE_INTERVAL = 3.0  # Segundos entre envíos
START_LAT = -17.7833  # Plaza 24 de Septiembre (Santa Cruz)
START_LON = -63.1821


def get_crc16(data):
    """Calcula CRC-ITU real para evitar rechazos en sesiones largas"""
    crc = 0xFFFF
    for byte in data:
        crc ^= byte
        for _ in range(8):
            if crc & 0x0001:
                crc = (crc >> 1) ^ 0x8408
            else:
                crc >>= 1
    return ~crc & 0xFFFF


def create_gt06_packet(proto, content, serial_num):
    start = b"\x78\x78"
    stop = b"\x0d\x0a"
    serial_bytes = struct.pack(">H", serial_num)
    length_val = 1 + len(content) + 2 + 2  # Proto + Content + Serial + ErrorCheck
    length = bytes([length_val])
    proto_byte = bytes([proto])

    body = length + proto_byte + content + serial_bytes

    # Calculamos CRC real del cuerpo (Length + Proto + Content + Serial)
    crc_val = get_crc16(body)
    crc_bytes = struct.pack(">H", crc_val)

    return start + body + crc_bytes + stop


def get_imei_bytes(imei_str):
    if len(imei_str) % 2 != 0:
        imei_str = "0" + imei_str
    return bytes.fromhex(imei_str)


def build_location_packet(lat, lon, speed, course, satellites, serial):
    # 1. FECHA ACTUAL (Importante para modo LIVE)
    now = datetime.utcnow()
    year = now.year - 2000
    datetime_bytes = struct.pack(
        "BBBBBB", year, now.month, now.day, now.hour, now.minute, now.second
    )

    # 2. SATÉLITES (Simulación de intensidad)
    # Primer nibble: Longitud GPS (fijo C o F), Segundo nibble: Cantidad Sats
    # Ejemplo: 0xf9 -> 9 satélites
    sat_byte = bytes([0xF0 | (satellites & 0x0F)])

    # 3. LATITUD (GT06 format)
    lat_val = int(abs(lat) * 1800000)
    lat_bytes = struct.pack(">I", lat_val)

    # 4. LONGITUD
    lon_val = int(abs(lon) * 1800000)
    lon_bytes = struct.pack(">I", lon_val)

    # 5. VELOCIDAD
    speed_byte = bytes([int(speed)])

    # 6. CURSO Y ESTADO (Bolivia: Sur y Oeste)
    # Bit 10 (0x0400): 0=Sur (Lat negativa)
    # Bit 11 (0x0800): 1=Oeste (Lon negativa)
    status_flags = 0x0800

    # Si la velocidad es > 0, asumimos GPS tracking ON (bit 12)
    if speed > 0:
        status_flags |= 0x1000

    course_combined = status_flags | (int(course) & 0x03FF)
    course_bytes = struct.pack(">H", course_combined)

    content = (
        datetime_bytes + sat_byte + lat_bytes + lon_bytes + speed_byte + course_bytes
    )

    # Protocolo 0x12 = Location Data
    return create_gt06_packet(0x12, content, serial)


def simulate_movement(lat, lon, speed, course):
    """Calcula nueva posición basada en velocidad y rumbo (Física básica)"""

    # Convertir a metros por segundo
    speed_ms = speed * 0.27778

    # Distancia recorrida en este intervalo
    distance_m = speed_ms * UPDATE_INTERVAL

    # Fórmula de Haversine inversa simplificada para distancias cortas
    # 1 grado latitud ~= 111,320 metros
    # 1 grado longitud ~= 111,320 * cos(lat) metros

    delta_lat = (distance_m * math.cos(math.radians(course))) / 111320
    delta_lon = (distance_m * math.sin(math.radians(course))) / (
        111320 * math.cos(math.radians(lat))
    )

    new_lat = lat + delta_lat
    new_lon = lon + delta_lon

    return new_lat, new_lon


def main():
    print(f"🚗 SIMULADOR DE RUTA GT06 - SANTA CRUZ")
    print(f"📡 Target: {HOST}:{PORT} | IMEI: {IMEI_TARGET}")

    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        s.connect((HOST, PORT))
        print("✅ Conexión TCP establecida.")

        # LOGIN
        imei_bytes = get_imei_bytes(IMEI_TARGET)
        login_packet = create_gt06_packet(0x01, imei_bytes, 1)
        s.sendall(login_packet)
        print("🔑 Login enviado.")
        time.sleep(1)

        # ESTADO INICIAL DEL VEHÍCULO
        current_lat = START_LAT
        current_lon = START_LON
        current_speed = 0
        current_course = random.randint(0, 360)
        current_sats = 10
        serial = 2

        print("\n🚀 Iniciando conducción automática...")
        print("Presiona CTRL+C para detener.\n")

        while True:
            # 1. LÓGICA DE CONDUCCIÓN (IA Simple)

            # Cambiar velocidad (Acelerar/Frenar)
            accel = random.uniform(-5, 10)  # Tendencia a acelerar suave
            current_speed += accel

            # Límites de velocidad
            if current_speed > 80:
                current_speed = 80  # Máximo urbano
            if current_speed < 0:
                current_speed = 0  # Detenido

            # Cambiar dirección (Girar)
            # Solo giramos si nos movemos
            if current_speed > 2:
                turn = random.uniform(-15, 15)  # Giro suave
                current_course = (current_course + turn) % 360

            # Calcular nueva posición física
            current_lat, current_lon = simulate_movement(
                current_lat, current_lon, current_speed, current_course
            )

            # Variar satélites (Simular túneles o árboles)
            # 90% probabilidad de tener buena señal, 10% de baja señal
            if random.random() > 0.9:
                current_sats = random.randint(0, 5)  # Mala señal
            else:
                current_sats = random.randint(6, 12)  # Buena señal

            # 2. CONSTRUIR Y ENVIAR
            packet = build_location_packet(
                current_lat,
                current_lon,
                current_speed,
                current_course,
                current_sats,
                serial,
            )
            s.sendall(packet)

            # 3. LOG EN CONSOLA (Para que veas qué pasa)
            status_icon = "🟢" if current_speed > 0 else "🔴"
            sat_icon = "📡" if current_sats > 5 else "⚠️"

            print(
                f"{status_icon} Spd: {int(current_speed)}km/h | 🧭 {int(current_course)}° | {sat_icon} Sats: {current_sats} | Lat: {current_lat:.5f} Lon: {current_lon:.5f}"
            )

            serial += 1
            time.sleep(UPDATE_INTERVAL)

    except KeyboardInterrupt:
        print("\n🛑 Simulación detenida por el usuario.")
    except Exception as e:
        print(f"\n❌ Error de conexión: {e}")
    finally:
        s.close()


if __name__ == "__main__":
    main()
