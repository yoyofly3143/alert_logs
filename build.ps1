# build.ps1
# PowerShell script to cross-compile alert-webhook for Linux (CentOS 7)

$ProjectRoot = Get-Location
$BuildDir = Join-Path $ProjectRoot "build"
$BinaryName = "alert-webhook-linux-amd64"

# Create build directory
if (-not (Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir | Out-Null
    Write-Host "Created build directory: $BuildDir" -ForegroundColor Cyan
}

# Set environment variables for Linux compilation
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"  # Static linking for better compatibility (important for CentOS 7)

Write-Host "Compiling for Linux/amd64..." -ForegroundColor Yellow
go build -ldflags="-s -w" -o (Join-Path $BuildDir $BinaryName) main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Compilation successful!" -ForegroundColor Green
    Write-Host "Output: $(Join-Path $BuildDir $BinaryName)"
    
    # Copy required assets
    Write-Host "Copying assets to build directory..." -ForegroundColor Yellow
    
    $Assets = @("templates", "static")
    foreach ($Asset in $Assets) {
        $Source = Join-Path $ProjectRoot $Asset
        $Dest = Join-Path $BuildDir $Asset
        if (Test-Path $Source) {
            if (Test-Path $Dest) { Remove-Item -Recurse -Force $Dest }
            Copy-Item -Recurse $Source $Dest
        }
    }
    
    # Copy .env.example
    if (Test-Path (Join-Path $ProjectRoot ".env")) {
        Copy-Item (Join-Path $ProjectRoot ".env") (Join-Path $BuildDir ".env")
    }
    
    Write-Host "Build complete! You can now transfer the 'build' folder to your CentOS 7 server." -ForegroundColor Green
} else {
    Write-Host "Compilation failed!" -ForegroundColor Red
}
