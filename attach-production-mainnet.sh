ODEX_ENABLE_TLS=true; \
ODEX_MONGODB_SHARD_URL_1=odexcluster0-shard-00-00-xzynf.mongodb.net:27017; \
ODEX_MONGODB_SHARD_URL_2=odexcluster0-shard-00-01-xzynf.mongodb.net:27017; \
ODEX_MONGODB_SHARD_URL_3=odexcluster0-shard-00-02-xzynf.mongodb.net:27017; \
fresh

# go run --race main.go