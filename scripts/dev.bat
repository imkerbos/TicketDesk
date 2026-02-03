@echo off
REM TicketDesk Development Server for Windows
REM Starts both backend (Air) and frontend (Vite) with hot reload

setlocal enabledelayedexpansion

echo.
echo ========================================
echo   TicketDesk Development Server
echo ========================================
echo.

REM Check dependencies
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed
    exit /b 1
)

where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Node.js is not installed
    exit /b 1
)

where air >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARN] Air is not installed, installing...
    go install github.com/air-verse/air@latest
)

REM Install frontend dependencies if needed
if not exist "web\node_modules" (
    echo [INFO] Installing frontend dependencies...
    cd web
    call npm install
    cd ..
)

echo.
echo [INFO] Starting backend server (Air hot reload)...
start "TicketDesk Backend" cmd /c "air -c .air.toml"

timeout /t 2 /nobreak >nul

echo [INFO] Starting frontend server (Vite hot reload)...
start "TicketDesk Frontend" cmd /c "cd web && npm run dev"

echo.
echo ========================================
echo   Development servers started!
echo ========================================
echo.
echo Backend API:  http://localhost:10010
echo Frontend:     http://localhost:3100
echo API Docs:     http://localhost:10010/swagger/index.html
echo.
echo Press Ctrl+C in each window to stop
echo.

pause
