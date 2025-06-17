package redis

//go:generate mockgen -destination=mock_universal_client.go -package=redis github.com/redis/go-redis/v9 UniversalClient
