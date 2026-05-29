
# Script para compilar version Windows desde Windows
Write-Host "Compilando para Windows (amd64)..."

$original_goos = $env:GOOS
$original_goarch = $env:GOARCH

$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Crear directorio bin si no existe
if (!(Test-Path -Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

go build -o bin/motos-socket.exe ./cmd/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build exitoso: bin/motos-socket.exe"
} else {
    Write-Host "❌ Error en el build"
}

# Restaurar variables
$env:GOOS = $original_goos
$env:GOARCH = $original_goarch
