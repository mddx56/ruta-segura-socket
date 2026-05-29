# Socket Server con Monitor WebSocket

Servidor TCP que recibe mensajes de clientes y los transmite en tiempo real a través de WebSocket. También envía todos los mensajes a un endpoint de API para registro.

## 🏗️ Arquitectura

El proyecto sigue una arquitectura limpia con separación de responsabilidades:

```archivos
motos-socket/
├── cmd/
│   ├── main.go          # Punto de entrada del servidor
│   └── monitor.html     # Interfaz web de monitoreo
├── internal/
│   ├── config/
│   │   └── config.go    # Configuración y variables de entorno
│   ├── handlers/
│   │   ├── health.go    # Handler del endpoint /health
│   │   ├── websocket.go # Handler de WebSocket
│   │   └── monitor.go   # Handler del monitor HTML
│   ├── routes/
│   │   └── routes.go    # Configuración de rutas
│   └── services/
│       ├── hub.go       # Servicio de Hub WebSocket
│       ├── hub_helpers.go
│       └── logger.go    # Servicio de logging a API
├── .env                 # Configuración
├── go.mod
└── README.md
```

## ✨ Características

- ✅ Servidor TCP que acepta conexiones de clientes
- ✅ Monitor web en tiempo real vía WebSocket
- ✅ Envío automático de mensajes a endpoint de API
- ✅ **Endpoint de Health Check** (`/health`)
- ✅ Framework **Gin** para HTTP
- ✅ Arquitectura limpia con separación de capas
- ✅ Interfaz de monitoreo en español

## 🔧 Configuración

Edita el archivo `.env`:

```env
SS_HOST=0.0.0.0              # Host del servidor TCP
SS_PORT=8080                 # Puerto del servidor TCP
WS_PORT=6060                 # Puerto del monitor web
LOG_URL=http://localhost:8888/api/log-socket  # Endpoint de API
API_KEY=dd9sad709asyd8y      # API Key para autenticación
```

## 🚀 Ejecución

### Desarrollo (Windows)

```bash
go run cmd/main.go
```

### Compilar para Windows

```bash
go build -o server.exe cmd/main.go
```

### Compilar para Linux

```bash
$env:GOOS='linux'; $env:GOARCH='amd64'; go build -o server-linux cmd/main.go
```

## 📡 Endpoints HTTP

### Monitor Web

- **URL**: `http://localhost:6060/`
- **Descripción**: Interfaz web para monitorear mensajes en tiempo real

### WebSocket

- **URL**: `ws://localhost:6060/ws`
- **Descripción**: Endpoint WebSocket para recibir mensajes en tiempo real

### Health Check

- **URL**: `http://localhost:6060/health`
- **Método**: GET
- **Descripción**: Verifica si el servidor está corriendo

**Respuesta de ejemplo:**

```json
{
  "status": "ok",
  "message": "Socket server está corriendo",
  "tcp_server": "0.0.0.0:8080",
  "websocket_port": "6060",
  "active_clients": 2,
  "log_endpoint": "http://localhost:8888/api/log-socket"
}
```

## 🔌 Servidor TCP

El servidor TCP escucha en el puerto configurado (por defecto 8080) y:

1. Recibe mensajes de clientes
2. Responde con "ok 200!"
3. Transmite el mensaje a todos los clientes WebSocket conectados
4. Envía el mensaje al endpoint de API para registro

## 📤 API Endpoint

Todos los mensajes recibidos se envían automáticamente a:

**POST** `http://localhost:8888/api/log-socket`

**Headers:**

```c
Content-Type: application/json
X-API-Key: dd9sad709asyd8y
```

**Body:**

```json
{
  "payload": "mensaje recibido del cliente"
}
```

## 🧪 Ejemplo de Cliente de Prueba

```python
import socket

HOST = 'localhost'
PORT = 8080

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
    s.connect((HOST, PORT))
    s.sendall(b"Test message")
    data = s.recv(1024)
    print(f"Received: {data.decode()}")
```

## 🐧 Despliegue en Servidor Ubuntu

1. Compila para Linux:

   ```bash
   $env:GOOS='linux'; $env:GOARCH='amd64'; go build -o server-linux cmd/main.go
   ```

2. Sube al servidor:
   - `server-linux`
   - `.env`

3. Da permisos de ejecución:

   ```bash
   chmod +x server-linux
   ```

4. Ejecuta:

   ```bash
   ./server-linux
   ```

## 📦 Dependencias

- `github.com/gin-gonic/gin` - Framework HTTP
- `github.com/gorilla/websocket` - WebSocket support
- `github.com/joho/godotenv` - Environment variables

Instalar con:

```bash
go mod tidy
```

## 🏛️ Arquitectura de Código

### Handlers

Manejan las peticiones HTTP y WebSocket:

- `health.go`: Endpoint de health check
- `websocket.go`: Manejo de conexiones WebSocket
- `monitor.go`: Servir la interfaz HTML

### Services

Lógica de negocio:

- `hub.go`: Gestión de clientes WebSocket y broadcasting
- `logger.go`: Envío de logs a la API externa

### Routes

Configuración centralizada de todas las rutas HTTP

### Config

Carga y gestión de variables de entorno

---

**Nota**: Después de hacer cambios en el código, recuerda reiniciar el servidor para que los cambios surtan efecto.
