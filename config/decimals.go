package config

import (
	"strconv"
	"strings"
)

func contarDecimales(num float64) int {
	// Convertimos el número a cadena de texto
	numStr := strconv.FormatFloat(num, 'f', -1, 64)

	// Dividimos la cadena por el punto decimal
	parts := strings.Split(numStr, ".")

	// Si hay una parte decimal, devolvemos su longitud, de lo contrario, devolvemos 0
	if len(parts) == 2 {
		return len(parts[1])
	}

	return 0
}

func validatePositionDecimals(lat float64, lng float64, cant int) bool {
	return contarDecimales(lat) >= cant && contarDecimales(lng) >= cant
}
