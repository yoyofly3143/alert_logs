# build.ps1
# PowerShell script to cross-compile alert-webhook for Linux (CentOS 7)

$ProjectRoot = Get-Location
$BuildDir = Join-Path $ProjectRoot "release"
$BinaryName = "alert-webhook"

# Create build directory
if (Test-Path $BuildDir) {
    Remove-Item -Recurse -Force $BuildDir
}
New-Item -ItemType Directory -Path $BuildDir | Out-Null
Write-Host "Created release directory: $BuildDir" -ForegroundColor Cyan

# Set environment variables for Linux compilation
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"  # Static linking for better compatibility (important for CentOS 7)

Write-Host "Compiling for Linux/amd64..." -ForegroundColor Yellow
# -ldflags="-s -w" reduces binary size
go build -ldflags="-s -w" -o (Join-Path $BuildDir $BinaryName) main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Compilation successful!" -ForegroundColor Green
    
    # Copy required assets
    Write-Host "Copying assets to release directory..." -ForegroundColor Yellow
    
    $Assets = @("templates", "static")
    foreach ($Asset in $Assets) {
        $Source = Join-Path $ProjectRoot $Asset
        $Dest = Join-Path $BuildDir $Asset
        if (Test-Path $Source) {
            Copy-Item -Recurse $Source $Dest
        }
    }
    
    # Copy .env
    if (Test-Path (Join-Path $ProjectRoot ".env")) {
        Copy-Item (Join-Path $ProjectRoot ".env") (Join-Path $BuildDir ".env")
        Write-Host "Copied .env" -ForegroundColor Gray
    }

    Write-Host "Build complete! The 'release' folder is ready for CentOS 7." -ForegroundColor Green
    Write-Host "Location: $BuildDir" -ForegroundColor White
} else {
    Write-Host "Compilation failed!" -ForegroundColor Red
    exit 1
}
