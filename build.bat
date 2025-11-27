@echo off
echo === Building LockScreen Sync ===
echo.

cd /d "%~dp0"

echo Downloading dependencies...
go mod tidy

echo.
echo Building executable...
go build -ldflags="-H windowsgui -s -w" -o LockScreenSync.exe .

if %errorlevel% equ 0 (
    echo.
    echo [OK] Build successful!
    echo Output: LockScreenSync.exe
    echo.
    echo NOTE: Run as Administrator for full functionality!
) else (
    echo.
    echo [ERROR] Build failed!
)

pause
