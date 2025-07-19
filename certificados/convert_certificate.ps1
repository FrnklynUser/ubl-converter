# Script para convertir certificado PFX con algoritmo de digest no soportado
# Convierte el certificado a un formato compatible con Go

param(
    [Parameter(Mandatory=$true)]
    [string]$InputPfx,
    
    [Parameter(Mandatory=$true)]
    [string]$Password,
    
    [Parameter(Mandatory=$false)]
    [string]$OutputPfx = ""
)

# Si no se especifica archivo de salida, usar el mismo nombre con sufijo "_converted"
if ($OutputPfx -eq "") {
    $baseName = [System.IO.Path]::GetFileNameWithoutExtension($InputPfx)
    $directory = [System.IO.Path]::GetDirectoryName($InputPfx)
    $OutputPfx = Join-Path $directory "$baseName`_converted.pfx"
}

Write-Host "Convirtiendo certificado PFX..." -ForegroundColor Green
Write-Host "Archivo de entrada: $InputPfx" -ForegroundColor Yellow
Write-Host "Archivo de salida: $OutputPfx" -ForegroundColor Yellow

try {
    # Verificar que OpenSSL esté disponible
    $opensslPath = Get-Command openssl -ErrorAction SilentlyContinue
    if (-not $opensslPath) {
        Write-Host "Error: OpenSSL no está instalado o no está en el PATH" -ForegroundColor Red
        Write-Host "Por favor instala OpenSSL desde: https://slproweb.com/products/Win32OpenSSL.html" -ForegroundColor Yellow
        exit 1
    }

    # Crear archivo temporal para PEM
    $tempPem = [System.IO.Path]::GetTempFileName() + ".pem"
    
    Write-Host "Paso 1: Extrayendo certificado y clave privada a formato PEM..." -ForegroundColor Cyan
    
    # Convertir PFX a PEM (sin cifrado)
    $convertToPemArgs = @(
        "pkcs12",
        "-in", $InputPfx,
        "-out", $tempPem,
        "-nodes",
        "-passin", "pass:$Password"
    )
    
    $result = & openssl @convertToPemArgs 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error al convertir a PEM: $result" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "Paso 2: Creando nuevo PFX con algoritmos compatibles..." -ForegroundColor Cyan
    
    # Convertir PEM de vuelta a PFX con algoritmos compatibles (sin legacy provider)
    $convertToPfxArgs = @(
        "pkcs12",
        "-export",
        "-in", $tempPem,
        "-out", $OutputPfx,
        "-passout", "pass:$Password",
        "-keypbe", "PBE-SHA1-3DES",
        "-certpbe", "PBE-SHA1-3DES",
        "-macalg", "SHA1"
    )
    
    $result = & openssl @convertToPfxArgs 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error al crear nuevo PFX: $result" -ForegroundColor Red
        Write-Host "Intentando método alternativo..." -ForegroundColor Yellow
        
        # Método alternativo sin especificar algoritmos específicos
        $convertToPfxArgs2 = @(
            "pkcs12",
            "-export",
            "-in", $tempPem,
            "-out", $OutputPfx,
            "-passout", "pass:$Password"
        )
        
        $result2 = & openssl @convertToPfxArgs2 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Error en método alternativo: $result2" -ForegroundColor Red
            exit 1
        }
    }
    
    # Limpiar archivo temporal
    Remove-Item $tempPem -Force -ErrorAction SilentlyContinue
    
    Write-Host "¡Conversión completada exitosamente!" -ForegroundColor Green
    Write-Host "Nuevo certificado guardado en: $OutputPfx" -ForegroundColor Green
    Write-Host ""
    Write-Host "Ahora puedes usar el certificado convertido en tu aplicación Go." -ForegroundColor Yellow
    
} catch {
    Write-Host "Error durante la conversión: $_" -ForegroundColor Red
    exit 1
}
