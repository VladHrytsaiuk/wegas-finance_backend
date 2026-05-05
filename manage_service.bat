@echo off
setlocal
cd /d "%~dp0"
title WeGaS Finance Service Manager
color 1F

:: --- 1. ПЕРЕВІРКА ПРАВ АДМІНІСТРАТОРА ---
net session >nul 2>&1
if %errorLevel% == 0 (
    goto :MENU
) else (
    color 4F
    echo.
    echo [ERROR] Bida... Nemaye prav Administratora!
    echo.
    echo Bud laska, natysnit pravoyu knopkoyu myshi na tsey fail
    echo i vyberit "Run as Administrator" (Zapustyty vid imeni administratora).
    echo.
    pause
    exit
)

:MENU
cls
echo ========================================================
echo           WEGAS FINANCE SERVICE MANAGER
echo ========================================================
echo.
echo    [1] INSTALL Service   (Vstanovyty sluzhbu)
echo    [2] START Service     (Zapustyty)
echo    [3] STOP Service      (Zupynyty)
echo    [4] RESTART Service   (Perezapustyty - dlya onovlennya .env)
echo    [5] UNINSTALL Service (Vydalyty sluzhbu)
echo    [6] STATUS Check      (Pereviryty stan)
echo.
echo    [0] EXIT
echo.
echo ========================================================
set /p choice="Vyberit diyu (0-6): "

if "%choice%"=="1" goto INSTALL
if "%choice%"=="2" goto START
if "%choice%"=="3" goto STOP
if "%choice%"=="4" goto RESTART
if "%choice%"=="5" goto UNINSTALL
if "%choice%"=="6" goto STATUS
if "%choice%"=="0" exit
goto MENU

:INSTALL
cls
echo Installing WeGaS Finance Service...
.\wegas-finance.exe install
if %errorlevel% neq 0 ( color 4F & echo [ERROR] Failed to install. ) else ( color 2F & echo [OK] Installed successfully. )
pause
color 1F
goto MENU

:START
cls
echo Starting Service...
.\wegas-finance.exe start
if %errorlevel% neq 0 ( color 4F & echo [ERROR] Failed to start. Check logs. ) else ( color 2F & echo [OK] Service started. )
pause
color 1F
goto MENU

:STOP
cls
echo Stopping Service...
.\wegas-finance.exe stop
if %errorlevel% neq 0 ( color 4F & echo [ERROR] Failed to stop. ) else ( color 2F & echo [OK] Service stopped. )
pause
color 1F
goto MENU

:RESTART
cls
color 0E
echo --- RESTARTING SERVICE ---
echo 1. Stopping...
.\wegas-finance.exe stop
echo.
echo 2. Waiting for port release (3 sec)...
timeout /t 3 /nobreak >nul
echo.
echo 3. Starting...
.\wegas-finance.exe start
echo.
if %errorlevel% neq 0 ( color 4F & echo [ERROR] Restart failed. ) else ( color 2F & echo [OK] Service restarted successfully! )
pause
color 1F
goto MENU

:UNINSTALL
cls
echo Uninstalling Service...
.\wegas-finance.exe uninstall
if %errorlevel% neq 0 ( color 4F & echo [ERROR] Failed to uninstall. ) else ( color 2F & echo [OK] Service removed. )
pause
color 1F
goto MENU

:STATUS
cls
echo Checking Status via SC...
sc query WeGaSFinance
echo.
pause
goto MENU