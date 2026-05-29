import socket
import struct
import binascii
import time

HOST = "localhost"
# HOST= '161.35.242.117'
PORT = 5050


def create_gt06_packet(proto, content, serial_num):
    # GT06 General Structure:
    # Start(2) + Length(1) + Proto(1) + Content(N) + Serial(2) + CRC(2) + Stop(2)

    start = b"\x78\x78"
    stop = b"\x0d\x0a"

    serial_bytes = struct.pack(">H", serial_num)

    # Calculate Length byte: Proto(1) + Content(N) + Serial(2) + CRC(2)
    # Note: Length byte usually excludes itself and Start/Stop.
    length_val = 1 + len(content) + 2 + 2
    length = bytes([length_val])

    proto_byte = bytes([proto])

    body = length + proto_byte + content + serial_bytes

    # Fake CRC for request (Client side calculation omitted for simplicity unless server checks it)
    crc = b"\x00\x00"

    return start + body + crc + stop


def verify_response(data, expected_proto, expected_serial):
    print(f"   Recibido (HEX): {binascii.hexlify(data)}")

    if len(data) < 10:
        print("   [FAIL] Respuesta muy corta")
        return False

    if data[0] != 0x78 or data[1] != 0x78:
        print("   [FAIL] Cabecera incorrecta")
        return False

    proto_resp = data[3]
    if proto_resp != expected_proto:
        print(
            f"   [FAIL] Protocolo incorrecto. Esperado {hex(expected_proto)}, Recibido {hex(proto_resp)}"
        )
        return False

    # Response structure: 78 78 [Len] [Proto] [SerialH] [SerialL] [CRC] [Stop]
    # Serial starts at index 4 (0-based) for standard response
    resp_serial = struct.unpack(">H", data[4:6])[0]

    if resp_serial != expected_serial:
        print(
            f"   [FAIL] Serial incorrecto. Esperado {expected_serial}, Recibido {resp_serial}"
        )
        return False

    print("   [PASS] Respuesta válida")
    return True


def main():
    print(f"Conectando a {HOST}:{PORT}...")
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect((HOST, PORT))
        print("Conectado.\n")

        # --- TEST 1: LOGIN (0x01) ---
        print("--- TEST 1: LOGIN (0x01) ---")
        imei = bytes.fromhex("0102030405060708")  # 8 bytes
        serial_login = 100
        packet_login = create_gt06_packet(0x01, imei, serial_login)

        print(f"   Enviando Login: {binascii.hexlify(packet_login)}")
        s.sendall(packet_login)
        data = s.recv(1024)
        verify_response(data, 0x01, serial_login)

        time.sleep(1)

        # --- TEST 2: HEARTBEAT (0x13) ---
        print("\n--- TEST 2: HEARTBEAT (0x13) ---")
        # Heartbeat content example: Status(1) + Voltage(1) + Signal(1) + Language(2) = 5 bytes
        # Dummy content
        hb_content = b"\x01\x02\x03\x04\x05"
        serial_hb = 200
        packet_hb = create_gt06_packet(0x13, hb_content, serial_hb)

        print(f"   Enviando Heartbeat: {binascii.hexlify(packet_hb)}")
        s.sendall(packet_hb)
        data = s.recv(1024)
        verify_response(data, 0x13, serial_hb)

        time.sleep(1)

        # --- TEST 3: LOCATION (0x12) ---
        print("\n--- TEST 3: LOCATION (0x12) ---")
        # Structure: Date(6) + Sat(1) + Lat(4) + Lon(4) + Speed(1) + Course(2) = 18 bytes content

        # Date: 18-01-2026 12:00:00 (Example Time)
        # Year(1), Month(1), Day(1), Hour(1), Min(1), Sec(1)
        # Year 2026 -> 26 (0x1A)
        datetime_bytes = b"\x1a\x01\x12\x0c\x00\x00"

        # Satellites info (1 byte)
        sat_byte = b"\x05"

        # Lat: -17.8138004
        # Value = 17.8138004 * 1800000 = 32064840 (0x01E94548)
        # Hex: 01 E9 45 48
        lat_bytes = b"\x01\xe9\x45\x48"

        # Lon: -63.1707532
        # Value = 63.1707532 * 1800000 = 113707355 (0x06C6E95B)
        # Hex: 06 C6 E9 5B
        lon_bytes = b"\x06\xc6\xe9\x5b"

        # Speed: 60 km/h (1 byte)
        speed_byte = b"\x3c"

        # Course/Status (2 bytes)
        # We need South (Bit 10 = 0) and West (Bit 11 = 1).
        # Standard GT06 bits:
        # Bit 10: 1=North, 0=South. So we want 0.
        # Bit 11: 1=West, 0=East. So we want 1. (0x0800)
        # Let's set some course value + 0x0800.
        # Value: 0x0814 (West, South, Course 20)
        course_bytes = b"\x14\x08"  # WATCH OUT! struct.pack('>H') is Big Endian.
        # If we send raw bytes: 0x08 0x14.
        # Bytes[20] = 0x08 (0000 1000) -> Bit 11 is 1 (West), Bit 10 is 0 (South implied by absence of 0x04).
        # Wait, 0x08 is 0000 1000. 0x04 is 0000 0100.
        # If byte[20] is 0x08 (00001000):
        # Bit 15-8 are in byte[20].
        # 0x08 = 0000 1000
        # Bits: 15=0, 14=0, 13=0, 12=0, 11=1 (West), 10=0 (South), 9=0, 8=0.
        # So YES, 0x08 in the first byte means West and South.
        course_bytes = b"\x08\x14"

        loc_content = (
            datetime_bytes
            + sat_byte
            + lat_bytes
            + lon_bytes
            + speed_byte
            + course_bytes
        )

        serial_loc = 300
        packet_loc = create_gt06_packet(0x12, loc_content, serial_loc)

        print(f"   Enviando Location: {binascii.hexlify(packet_loc)}")
        s.sendall(packet_loc)
        print("   Paquete enviado (No se espera ACK inmediato para 0x12 en demo).")

    except ConnectionRefusedError:
        print(
            "ERROR: No se pudo conectar. Asegúrate de que el servidor Go esté corriendo."
        )
    except Exception as e:
        print(f"ERROR: {e}")


if __name__ == "__main__":
    main()
