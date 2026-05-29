package tcp

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/waltherx/motos-socket/internal/cache"
	"github.com/waltherx/motos-socket/internal/gps/gt06"
	"github.com/waltherx/motos-socket/internal/services"
	"github.com/waltherx/motos-socket/internal/worker"
	grpcclient "github.com/waltherx/motos-socket/pkg/grpc_client"
)

// ConnectionHandler manages the lifecycle of a single GPS device TCP connection.
type ConnectionHandler struct {
	deviceCache cache.DeviceCache
	posWorker   *worker.PositionWorker
	grpcClient  *grpcclient.GRPCClient
}

// NewConnectionHandler wires all dependencies for a connection handler.
func NewConnectionHandler(
	dc cache.DeviceCache,
	pw *worker.PositionWorker,
	gc *grpcclient.GRPCClient,
) *ConnectionHandler {
	return &ConnectionHandler{
		deviceCache: dc,
		posWorker:   pw,
		grpcClient:  gc,
	}
}

// Handle runs the read loop for a single GPS device until the connection closes
// or the context is cancelled (shutdown).
func (h *ConnectionHandler) Handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Per-connection state (battery and ignition carry over across packets)
	var sessionIMEI string
	var lastBattery int
	var lastIgnition bool

	buf := make([]byte, 1024)
	var remain []byte // Acumulador para lidiar con fragmentación TCP
	remoteAddr := conn.RemoteAddr().String()
	slog.Info("Conexión GPS nueva", "remote", remoteAddr)

	for {
		// Check for graceful shutdown before every read
		select {
		case <-ctx.Done():
			slog.Info("Cerrando conexión GPS por shutdown", "imei", sessionIMEI, "remote", remoteAddr)
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
		n, err := conn.Read(buf)
		if err != nil {
			imeiLog := sessionIMEI
			if imeiLog == "" {
				imeiLog = "NO_LOGIN"
			}
			slog.Info("GPS desconectado", "remote", remoteAddr, "imei", imeiLog, "reason", err.Error())
			return
		}
		if n == 0 {
			continue
		}

		// Acumular bytes recibidos
		remain = append(remain, buf[:n]...)

		// Procesar todos los paquetes completos que haya en el buffer acumulado
		for len(remain) >= 5 {
			// Buscar la cabecera válida de inicio de paquete (0x78 0x78)
			if remain[0] != 0x78 || remain[1] != 0x78 {
				// Si no es válida, descartar un byte y reintentar sincronizar el stream
				services.PacketsIgnored.WithLabelValues("invalid_header").Inc()
				remain = remain[1:]
				continue
			}

			// En GT06, el byte de longitud (índice 2) cuenta el tamaño del Protocolo, Datos, Serial y Checksum.
			// Tamaño total del paquete = Start(2) + LengthByte(1) + Length + Stop(2) = Length + 5
			packetLength := int(remain[2])
			totalSize := packetLength + 5

			// ¿El paquete está completo en el buffer?
			if len(remain) < totalSize {
				break // Faltan fragmentos, salir del loop y esperar el próximo Read()
			}

			// Extraer paquete completo
			packet := remain[:totalSize]
			
			// Procesar el paquete
			h.processPacket(packet, conn, &sessionIMEI, &lastBattery, &lastIgnition, remoteAddr)

			// Avanzar el buffer acumulador descartando el paquete ya procesado
			remain = remain[totalSize:]
		}
	}
}

// processPacket handles the business logic of a single, fully-formed GT06 packet.
func (h *ConnectionHandler) processPacket(
	data []byte,
	conn net.Conn,
	sessionIMEI *string,
	lastBattery *int,
	lastIgnition *bool,
	remoteAddr string,
) {
	// Protocol byte
	switch data[3] {

	// ── LOGIN ────────────────────────────────────────────────────────────
	case gt06.ProtocolLogin:
		services.PacketsReceived.WithLabelValues("login").Inc()
		imei, ok := gt06.ParseIMEI(data)
		if !ok {
			return
		}
		*sessionIMEI = imei
		if h.deviceCache.IsKnown(imei) {
			slog.Info("Login GPS exitoso", "imei", imei, "remote", remoteAddr)
		} else {
			slog.Warn("Login GPS no registrado en caché", "imei", imei, "remote", remoteAddr)
		}
		conn.Write(gt06.CreateAck(data))

	// ── HEARTBEAT ────────────────────────────────────────────────────────
	case gt06.ProtocolHeartbeat:
		services.PacketsReceived.WithLabelValues("heartbeat").Inc()
		hb, ok := gt06.ParseHeartbeat(data)
		if !ok {
			return
		}
		*lastBattery = hb.Battery
		*lastIgnition = hb.Ignition
		conn.Write(gt06.CreateAck(data))

	// ── LOCATION ─────────────────────────────────────────────────────────
	case gt06.ProtocolLocation:
		services.PacketsReceived.WithLabelValues("location").Inc()
		if *sessionIMEI == "" {
			services.PacketsIgnored.WithLabelValues("no_imei").Inc()
			return
		}
		loc, ok := gt06.ParseLocation(data, *lastBattery, *lastIgnition)
		if !ok {
			return
		}
		// Update local ignition state (ParseLocation may infer it from speed)
		*lastIgnition = loc.Ignition

		// Persist via gRPC through the worker pool
		task := worker.PositionTask{
			Client: h.grpcClient,
			IMEI:   *sessionIMEI,
			Packet: loc,
		}
		if !h.posWorker.Enqueue(task) {
			services.PositionAPIDropped.Inc()
			slog.Warn("Worker pool lleno o cerrado, omitiendo paquete", "imei", *sessionIMEI)
		}
	}
}
