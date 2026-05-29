import socket
import struct
import binascii
import time

HOST = "161.35.242.117"
PORT = 5050


def create_gt06_packet(proto, content, serial_num):
    # GT06 General Structure:
    # Start(2) + Length(1) + Proto(1) + Content(N) + Serial(2) + CRC(2) + Stop(2)

    start = b"\x78\x78"
    stop = b"\x0d\x0a"

    serial_bytes = struct.pack(">H", serial_num)

    # Length byte: Proto(1) + Content(N) + Serial(2) + CRC(2)
    length_val = 1 + len(content) + 2 + 2
    length = bytes([length_val])
    proto_byte = bytes([proto])
    body = length + proto_byte + content + serial_bytes
    crc = b"\x00\x00"

    return start + body + crc + stop


def main():
    print(f"Connecting to {HOST}:{PORT} to test TRASH data filtering...")
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect((HOST, PORT))
        s.settimeout(2.0)  # Set timeout to catch lack of response for trash
        print("Connected.\n")

        # --- TEST 1: SEND TRASH ---
        print("--- TEST 1: SEND TRASH DATA ---")
        trash_data = b"This is not a GT06 packet. Just garbage."
        print(f"   Sending Trash: {trash_data}")
        s.sendall(trash_data)

        try:
            data = s.recv(1024)
            if data:
                print(f"   [FAIL] Server responded to trash: {binascii.hexlify(data)}")
            else:
                print(
                    "   [PASS] Server closed connection without data (or just ignored)."
                )
        except socket.timeout:
            print("   [PASS] No response from server (Timeout), as expected.")

        time.sleep(1)

        # --- TEST 2: SEND TRASH HEX (Invalid Header) ---
        print("\n--- TEST 2: SEND INVALID HEX HEADER ---")
        # Starts with 79 79 instead of 78 78
        invalid_header = b"\x79\x79\x05\x01\x00\x01\x00\x00\x0d\x0a"
        print(f"   Sending Invalid Header: {binascii.hexlify(invalid_header)}")
        s.sendall(invalid_header)

        try:
            data = s.recv(1024)
            if data:
                print(
                    f"   [FAIL] Server responded to invalid header: {binascii.hexlify(data)}"
                )
            else:
                print("   [PASS] Server closed/ignored.")
        except socket.timeout:
            print("   [PASS] No response from server (Timeout), as expected.")

        time.sleep(1)

        # --- TEST 3: SEND VALID GT06 ---
        print("\n--- TEST 3: SEND VALID GT06 LOGIN ---")
        imei = bytes.fromhex("0102030405060708")
        serial = 123
        packet = create_gt06_packet(0x01, imei, serial)
        print(f"   Sending Valid GT06: {binascii.hexlify(packet)}")
        s.sendall(packet)

        try:
            data = s.recv(1024)
            if data:
                print(f"   [PASS] Server responded: {binascii.hexlify(data)}")
            else:
                print("   [FAIL] Server did not respond to valid packet.")
        except socket.timeout:
            print("   [FAIL] Server timeout on valid packet.")

        s.close()

    except ConnectionRefusedError:
        print("ERROR: Could not connect to server.")
    except Exception as e:
        print(f"ERROR: {e}")


if __name__ == "__main__":
    main()
