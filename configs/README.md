# Configuration Directory

This directory contains YAML configuration files for different environments and use cases.

## Configuration Files

### `app.yaml`
**Base configuration** - Defines the complete configuration schema with default values. This serves as:
- Documentation for all available configuration options
- Default fallback values
- Schema reference for environment-specific overrides

### `development.yaml`
**Development environment** - Optimized for local development:
- More permissive CORS settings
- Higher rate limits for testing
- Shorter cache TTLs to see changes quickly
- Debug logging enabled
- HTTP-only sessions for easier testing

### `production.yaml`
**Production environment** - Secure and optimized for production:
- Restrictive CORS settings
- Lower rate limits for security
- Longer cache TTLs for performance
- JSON logging for structured logs
- HTTPS-only sessions
- Enhanced security settings

### `testing.yaml`
**Testing environment** - Configured for running tests:
- Different ports to avoid conflicts
- Separate test databases
- Minimal logging to reduce noise
- Disabled external features (push notifications, etc.)
- Very short cache TTLs for test isolation

## Usage

Currently, these configuration files serve as **documentation and reference**. The application uses environment variables for configuration, which provides:
- Better security (secrets not in files)
- Easier deployment (container-friendly)
- Platform flexibility (12-factor app compliance)

### Environment Variable Mapping

The YAML structure maps to environment variables as follows:

```yaml
# YAML Configuration
database:
  host: "localhost"
  port: 5432
  name: "matching_db"
```

```bash
# Environment Variables
DB_HOST=localhost
DB_PORT=5432
DB_NAME=matching_db
```

### Future Enhancements

These configuration files can be extended to support:

1. **Configuration Loading**: Parse YAML files and merge with environment variables
2. **Multi-Environment Deployment**: Load different configs based on `ENV` variable
3. **Validation**: Ensure all required configuration is present
4. **Hot Reloading**: Update configuration without restarting the service

### Example Implementation

```go
// Future enhancement - configuration loading
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    // ... other config sections
}

func LoadConfig(env string) (*Config, error) {
    // Load base config from app.yaml
    baseConfig, err := loadYAML("configs/app.yaml")
    if err != nil {
        return nil, err
    }
    
    // Override with environment-specific config
    envConfig, err := loadYAML(fmt.Sprintf("configs/%s.yaml", env))
    if err == nil {
        baseConfig = mergeConfigs(baseConfig, envConfig)
    }
    
    // Override with environment variables
    return overrideWithEnvVars(baseConfig), nil
}
```

## Configuration Reference

### Key Sections

- **Server**: HTTP server settings, timeouts, concurrency
- **Database**: PostgreSQL connection and pool settings
- **Redis**: Redis connection, caching, and session storage
- **Session**: Cookie settings, security, and lifetime
- **JWT**: Token settings and security
- **Rate Limiting**: Request throttling and abuse prevention
- **CORS**: Cross-origin request policies
- **AWS**: S3 integration for image storage
- **Cache**: TTL settings for different data types
- **Matching**: Algorithm parameters and constraints
- **Logging**: Output format and verbosity
- **Features**: Feature flags for gradual rollouts

### Security Considerations

- **Secrets**: Never store secrets in configuration files
- **Environment Variables**: Use for all sensitive data
- **Production Settings**: Always use HTTPS, secure cookies, and restrictive CORS
- **Rate Limiting**: Essential for preventing abuse
- **Logging**: Be careful not to log sensitive information

### Performance Tuning

- **Cache TTLs**: Balance between performance and data freshness
- **Connection Pools**: Tune based on expected load
- **Timeouts**: Set appropriate values for your infrastructure
- **Compression**: Enable for better bandwidth usage
- **Rate Limits**: Balance between user experience and resource protection