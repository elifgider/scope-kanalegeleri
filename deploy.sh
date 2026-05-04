#!/usr/bin/env bash
set -e

app="scope-kanalegeleri"
host=$1
public_port="8080" 

copy="app_binary Dockerfile templates public .env docker-compose.yml"

if [ -z "$host" ]; then
    read -p "Sunucu IP (host): " host
fi

echo "🚀 Dağıtım başlıyor: $host"
GOOS=linux GOARCH=amd64 go build -o app_binary ./cmd/app/main.go

ssh root@$host "mkdir -p /home/$app"
rsync -avz --delete $copy root@$host:/home/$app/

ssh root@$host "cd /home/$app; docker compose down || true"
ssh root@$host "cd /home/$app; docker compose build --pull"
ssh root@$host "cd /home/$app; docker compose up -d"

ssh root@$host 'docker update --restart unless-stopped $(docker ps -q)'

rm app_binary

echo "✅ Tamamlandı: http://$host:$public_port"