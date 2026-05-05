@echo off
set SERVICE_NAME=WeGaSFinance
set EXE_NAME=wegas_finance.exe
set NEW_EXE_NAME=wegas_finance_new.exe

echo ==========================================
echo      WeGaS Finance Update Utility
echo ==========================================

:: 1. Перевіряємо, чи є новий файл
if not exist %NEW_EXE_NAME% (
    echo [ERROR] File %NEW_EXE_NAME% not found!
    echo Put the new file next to this script.
    pause
    exit /b
)

:: 2. Зупиняємо службу (net stop чекає повної зупинки)
echo [1/4] Stopping Windows Service...
net stop %SERVICE_NAME%

:: Якщо служба не була запущена, net stop видасть помилку, але це ок.
:: Даємо ще 2 секунди про всяк випадок
timeout /t 2 /nobreak > nul

:: 3. Робимо бекап і заміну
echo [2/4] Updating files...

:: Видаляємо старий бекап, якщо був
if exist %EXE_NAME%.bak del %EXE_NAME%.bak

:: Перейменовуємо поточний файл у .bak (це надійніше, ніж видалення)
ren %EXE_NAME% %EXE_NAME%.bak

:: Ставимо новий файл
ren %NEW_EXE_NAME% %EXE_NAME%

echo [SUCCESS] File replaced.

:: 4. Запускаємо службу
echo [3/4] Starting Service...
net start %SERVICE_NAME%

echo ==========================================
echo      Update Completed Successfully!
echo ==========================================
pause