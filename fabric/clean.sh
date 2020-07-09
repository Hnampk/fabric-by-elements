docker-compose down
docker rm $(docker ps -aq)
docker volume prune

# rm -r ../node-sdk/fabric-client-kv-org*
