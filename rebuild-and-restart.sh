#!/bin/bash

# Script to rebuild and restart Oracle Engine with new email verification endpoints

echo "🔄 Stopping Oracle Engine containers..."
docker-compose down

echo "🏗️  Rebuilding Docker image with new code..."
docker-compose build --no-cache

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Starting containers..."
    docker-compose up -d
    
    echo "⏳ Waiting for services to start..."
    sleep 10
    
    echo "🧪 Testing health endpoint..."
    curl -s http://localhost:8000/api/health
    
    echo ""
    echo "🧪 Testing new email verification endpoint..."
    curl -s -X POST http://localhost:8000/api/auth/register/initiate \
      -H "Content-Type: application/json" \
      -d '{"email": "test@example.com"}' | jq .
    
    echo ""
    echo "✅ Oracle Engine restarted successfully!"
    echo "📚 Check logs with: docker-compose logs -f oracle"
else
    echo "❌ Build failed. Check your network connection and try again."
    echo "   Tip: Make sure you can reach docker.io"
    exit 1
fi

