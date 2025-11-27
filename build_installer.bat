@echo off
echo ============================================
echo   Building LockScreen Sync Installer
echo ============================================
echo.

cd /d "%~dp0"

:: Перевірка наявності exe файлу
if not exist "LockScreenSync.exe" (
    echo [ERROR] LockScreenSync.exe not found!
    echo Run build.bat first to compile the application.
    pause
    exit /b 1
)

:: Перевірка наявності іконки
if not exist "icon.ico" (
    echo [WARNING] icon.ico not found, creating default icon...
    powershell -Command "& {Add-Type -AssemblyName System.Drawing; $bmp = New-Object System.Drawing.Bitmap(32,32); $g = [System.Drawing.Graphics]::FromImage($bmp); $g.Clear([System.Drawing.Color]::FromArgb(66,133,244)); $g.FillRectangle([System.Drawing.Brushes]::White, 8, 6, 16, 12); $g.FillRectangle([System.Drawing.Brushes]::Gray, 13, 18, 6, 4); $g.FillRectangle([System.Drawing.Brushes]::Gray, 10, 22, 12, 2); $bmp.Save('%~dp0icon_temp.bmp'); $g.Dispose(); $bmp.Dispose()}"

    :: Використовуємо ImageMagick якщо є, інакше просто копіюємо bmp
    where magick >nul 2>&1
    if %errorlevel% equ 0 (
        magick convert icon_temp.bmp icon.ico
        del icon_temp.bmp
    ) else (
        echo [INFO] ImageMagick not found, using placeholder icon
        copy /Y "C:\Windows\System32\shell32.dll,15" icon.ico >nul 2>&1
        if not exist "icon.ico" (
            echo Creating minimal icon...
            powershell -Command "& {[System.IO.File]::WriteAllBytes('%~dp0icon.ico', [byte[]](0,0,1,0,1,0,16,16,0,0,1,0,32,0,104,4,0,0,22,0,0,0,40,0,0,0,16,0,0,0,32,0,0,0,1,0,32,0,0,0,0,0,0,4,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0))}"
        )
    )
)

:: Створення папки для інсталятора
if not exist "installer" mkdir installer

:: Шлях до Inno Setup (стандартні місця)
set "ISCC="
if exist "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" set "ISCC=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
if exist "C:\Program Files\Inno Setup 6\ISCC.exe" set "ISCC=C:\Program Files\Inno Setup 6\ISCC.exe"
if exist "%LOCALAPPDATA%\Programs\Inno Setup 6\ISCC.exe" set "ISCC=%LOCALAPPDATA%\Programs\Inno Setup 6\ISCC.exe"

if "%ISCC%"=="" (
    echo.
    echo [ERROR] Inno Setup 6 not found!
    echo.
    echo Please install Inno Setup 6 from:
    echo https://jrsoftware.org/isdl.php
    echo.
    pause
    exit /b 1
)

echo.
echo Compiling installer with Inno Setup...
echo.

"%ISCC%" setup.iss

if %errorlevel% equ 0 (
    echo.
    echo ============================================
    echo   [OK] Installer created successfully!
    echo   Output: installer\LockScreenSync_Setup.exe
    echo ============================================
) else (
    echo.
    echo [ERROR] Failed to create installer!
)

echo.
pause
