# Oracle Engine

A real-time, reliable asset price oracle system that provides aggregated moving window algorithm to ensure stability and reduce manipulation. Built with Go, featuring multiple data sources, consensus mechanisms, and a RESTful API with Swagger documentation.

## Features

- **Multi-Source Data Aggregation**: Integrates with multiple price feed providers (Pyth, MonieRate, ExchangeRate, TwelveData, Fixer, CurrencyLayer, Moralis)
- **Consensus Mechanism**: Weighted voting system for price validation
- **Real-time Price Streaming**: Server-Sent Events (SSE) for live price updates
- **RESTful API**: Comprehensive API with Swagger documentation
- **Database Integration**: TimescaleDB for time-series data storage
- **Caching**: Redis for high-performance caching
- **Docker Support**: Containerized deployment with Docker Compose
- **Hot Reloading**: Development mode with Air for automatic code reloading

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Data Sources  │    │   Price Pool    │    │   Aggregator    │
│                 │    │                 │    │                 │
│ • Pyth         │───▶│ • Outlier       │───▶│ • Consensus     │
│ • MonieRate    │    │   Detection     │    │   Algorithm     │
│ • ExchangeRate │    │ • DLQ           │    │ • Weighted      │
│ • TwelveData   │    │                 │    │   Voting        │
│ • Fixer        │    └─────────────────┘    └─────────────────┘
│ • CurrencyLayer│                                    │
│ • Moralis      │                                    ▼
└─────────────────┘                         ┌─────────────────┐
                                            │   Relayer       │
                                            │                 │
                                            │ • Blockchain    │
                                            │   Integration   │
                                            └─────────────────┘
```

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Redis
- TimescaleDB (PostgreSQL 16+)

## Quick Start

### Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd oracle_engine
   ```

2. **Set up environment variables**
   ```bash
   cp copy.env .env
   # Edit .env with your configuration
   ```

3. **Generate Swagger documentation**
   ```bash
   make swag
   ```

4. **Run the application**
   ```bash
   make run
   ```

The application will be available at:
- API: http://localhost:8000
- Swagger UI: http://localhost:8000/swagger/
- Health Check: http://localhost:8000/api/health

### Development Mode

For development with hot reloading:

```bash
make dev
```

This will start the application in development mode with automatic code reloading using Air.

## API Endpoints

### Price Endpoints
- `GET /api/prices/last` - Get the last known price for an asset
- `GET /api/prices/stream` - Stream real-time price updates (SSE)
- `GET /api/prices/:id/audit` - Get price audit information

### Issuance Endpoints
- `POST /api/issuances` - Create a new price issuance
- `GET /api/issuances/:id` - Get issuance details

### Asset Endpoints
- `GET /api/assets` - Get list of supported assets

### Health Check
- `GET /api/health` - Health check endpoint

## Configuration

The application uses a `config.yaml` file for configuration. Key settings include:

- Database connection strings
- Redis configuration
- Asset definitions
- API keys for data providers
- Consensus parameters

## Development Workflow

### For Developers

1. **Local Development**
   ```bash
   # Start development environment
   make dev
   ```

2. **Before Pushing to Staging**
   ```bash
   # Generate swagger documentation locally
   make swag
   
   # Commit and push changes
   git add .
   git commit -m "Update API documentation"
   git push
   ```

3. **On Staging Server**
   ```bash
   # Pull latest changes
   git pull
   
   # Run in development mode (swagger docs already generated)
   make dev
   ```

### Important Note for Developers

**Always run `make swag` locally before pushing to staging!** This ensures that:
- Swagger documentation is up-to-date with your API changes
- The staging environment can run `make run` without needing to regenerate docs
- API documentation remains consistent across environments

### Available Make Commands

- `make run` - Run the application with Docker Compose
- `make dev` - Run in development mode with hot reloading
- `make build` - Build Docker containers
- `make swag` - Generate Swagger documentation
- `make swag-clean` - Clean Swagger documentation
- `make clean` - Clean up containers and temporary files
- `make up` - Start containers in detached mode
- `make down` - Stop and remove containers

## Data Sources

The oracle integrates with multiple price feed providers:

- **Pyth Network** - Decentralized price feeds
- **MonieRate** - Cryptocurrency price data
- **ExchangeRate-API** - Foreign exchange rates
- **TwelveData** - Financial market data
- **Fixer.io** - Currency conversion rates
- **CurrencyLayer** - Real-time exchange rates
- **Moralis** - Web3 data and APIs

## Consensus Algorithm

The system uses a weighted consensus algorithm that:
- Aggregates prices from multiple sources
- Applies outlier detection to remove anomalies
- Uses weighted voting based on source reliability
- Implements a moving window for stability
- Provides audit trails for transparency

## Database Schema

The application uses TimescaleDB (PostgreSQL extension) for time-series data:

- **Prices Table**: Stores historical price data with hypertables
- **Issuances Table**: Records price consensus results
- **Audit Table**: Tracks price changes and validations

## Monitoring & Logging

- Structured logging with Zap
- Health check endpoints
- Price audit trails
- Performance metrics

## Deployment

### Production Deployment

1. **Build the production image**
   ```bash
   make build
   ```

2. **Deploy with Docker Compose**
   ```bash
   make up
   ```

### Environment Variables

Required environment variables:
- `POSTGRES_USER` - Database username
- `POSTGRES_PASSWORD` - Database password
- `POSTGRES_DB` - Database name
- `REDIS_HOST` - Redis host
- `REDIS_PORT` - Redis port

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `make swag` to update documentation
5. Test your changes
6. Submit a pull request

## License

Apache 2.0 License

## Support

For support and questions:
- Email: ifalabstudio@gmail.com
- Website: https://ifalabs.com
