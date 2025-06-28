# Redis Database Usage

To prevent integration tests from wiping development data, we use different Redis databases:

## Database Assignments

- **DB 0** (default): Development environment
  - Used when running `make run` or `./bin/dnd-bot`
  - Contains your actual character data
  
- **DB 15**: Test environment
  - Used by `make test-integration`
  - Automatically cleared before/after tests
  - Safe to wipe without affecting dev data

## Running Tests Safely

Always use the Makefile commands:
```bash
make test-integration     # Automatically uses DB 15
make test-all            # Runs all tests with proper isolation
```

## Manual Testing

If you need to run tests manually, ensure you set the Redis URL:
```bash
REDIS_URL="redis://localhost:6379/15" go test ./... -tags=integration
```

## Checking Your Data

To verify which database has data:
```bash
# Check dev database (0)
redis-cli -n 0 DBSIZE

# Check test database (15)
redis-cli -n 15 DBSIZE
```

## Migration

If you accidentally ran tests against DB 0 and lost data, you might be able to recover from Redis persistence files or backups.