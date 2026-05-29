Write-Host "Generating Swagger docs..." -ForegroundColor Cyan
swag init -g cmd/main.go

if ($LASTEXITCODE -ne 0) {
    Write-Warning "Swagger generation failed! Continuing..."
}

Write-Host "Swagger UI : http://localhost:6060/swagger/index.html" -ForegroundColor Yellow
Write-Host "Starting Socket Server..." -ForegroundColor Green
go run cmd/main.go
