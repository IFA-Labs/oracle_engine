@echo off
REM Script to rebuild and restart Oracle Engine with new email verification endpoints

echo Stopping Oracle Engine containers...
docker-compose down

echo Building Docker image with new code...
docker-compose build --no-cache

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Starting containers...
    docker-compose up -d
    
    echo Waiting for services to start...
    timeout /t 10 /nobreak >nul
    
    echo Testing health endpoint...
    curl -s http://localhost:8000/api/health
    
    echo.
    echo Testing new email verification endpoint...
    curl -s -X POST http://localhost:8000/api/auth/register/initiate -H "Content-Type: application/json" -d "{\"email\": \"test@example.com\"}"
    
    echo.
    echo Oracle Engine restarted successfully!
    echo Check logs with: docker-compose logs -f oracle
) else (
    echo Build failed. Check your network connection and try again.
    echo Tip: Make sure you can reach docker.io
    exit /b 1
)

