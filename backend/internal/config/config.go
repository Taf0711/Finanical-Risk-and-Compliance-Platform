package config

import (
    "log"
    "os"
    "strconv"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    App      AppConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
    WS       WebSocketConfig
    Risk     RiskConfig
    Alert    AlertConfig
}

type AppConfig struct {
    Env  string
    Port string
    Name string
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

type JWTConfig struct {
    Secret string
    Expiry time.Duration
}

type WebSocketConfig struct {
    ReadBufferSize  int
    WriteBufferSize int
}

type RiskConfig struct {
    VARConfidenceLevel  float64
    VARTimeHorizon      int
    LiquidityThreshold  float64
    PositionLimitPercent float64
}

type AlertConfig struct {
    CleanupDays int
    BatchSize   int
}

func Load() (*Config, error) {
    err := godotenv.Load()
    if err != nil {
        log.Printf("Warning: .env file not found")
    }

    return &Config{
        App: AppConfig{
            Env:  getEnv("APP_ENV", "development"),
            Port: getEnv("APP_PORT", "8080"),
            Name: getEnv("APP_NAME", "Financial Risk Monitor"),
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "riskmonitor"),
            Password: getEnv("DB_PASSWORD", ""),
            DBName:   getEnv("DB_NAME", "financial_risk_db"),
            SSLMode:  getEnv("DB_SSL_MODE", "disable"),
        },
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnv("REDIS_PORT", "6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
            DB:       getEnvAsInt("REDIS_DB", 0),
        },
        JWT: JWTConfig{
            Secret: getEnv("JWT_SECRET", "your-secret-key"),
            Expiry: getEnvAsDuration("JWT_EXPIRY", "24h"),
        },
        WS: WebSocketConfig{
            ReadBufferSize:  getEnvAsInt("WS_READ_BUFFER_SIZE", 1024),
            WriteBufferSize: getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024),
        },
        Risk: RiskConfig{
            VARConfidenceLevel:   getEnvAsFloat("VAR_CONFIDENCE_LEVEL", 0.95),
            VARTimeHorizon:       getEnvAsInt("VAR_TIME_HORIZON", 1),
            LiquidityThreshold:   getEnvAsFloat("LIQUIDITY_THRESHOLD", 0.3),
            PositionLimitPercent: getEnvAsFloat("POSITION_LIMIT_PERCENT", 25.0),
        },
        Alert: AlertConfig{
            CleanupDays: getEnvAsInt("ALERT_CLEANUP_DAYS", 30),
            BatchSize:   getEnvAsInt("ALERT_BATCH_SIZE", 100),
        },
    }, nil
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    valueStr := getEnv(key, "")
    if value, err := strconv.Atoi(valueStr); err == nil {
        return value
    }
    return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
    valueStr := getEnv(key, "")
    if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
        return value
    }
    return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
    valueStr := getEnv(key, defaultValue)
    if value, err := time.ParseDuration(valueStr); err == nil {
        return value
    }
    duration, _ := time.ParseDuration(defaultValue)
    return duration
}