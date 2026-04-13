#!/usr/bin/env bash
set -e

read -rp "Sunucu IP: " SERVER_IP
read -rsp "SSH Şifresi: " SERVER_PASS && echo

SERVER_USER="elifadmin"
APP_ROOT="/opt/kanalegeleri-go"
LOCAL_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOCAL_GO_APP="$LOCAL_ROOT/go-app"

if ! command -v sshpass >/dev/null 2>&1; then
  echo "⚙️   sshpass bulunamadı, kuruluyor (Homebrew)..."
  brew install sshpass
fi

SSH="sshpass -p $SERVER_PASS ssh -tt -o StrictHostKeyChecking=no"

echo "🛠️   Hedef klasör hazırlanıyor..."
$SSH "$SERVER_USER@$SERVER_IP" "echo '$SERVER_PASS' | sudo -S mkdir -p '$APP_ROOT/go-app' '$APP_ROOT/public' && echo '$SERVER_PASS' | sudo -S chown -R '$SERVER_USER:$SERVER_USER' '$APP_ROOT'"

echo "📦  Go uygulaması kopyalanıyor..."
sshpass -p "$SERVER_PASS" rsync -az --delete \
  -e "ssh -o StrictHostKeyChecking=no" \
  --exclude='.git' \
  --exclude='uploads' \
  "$LOCAL_GO_APP/" "$SERVER_USER@$SERVER_IP:$APP_ROOT/go-app/"

echo "🖼️   Statik görseller kopyalanıyor..."
sshpass -p "$SERVER_PASS" rsync -az --delete \
  -e "ssh -o StrictHostKeyChecking=no" \
  "$LOCAL_ROOT/public/static/" "$SERVER_USER@$SERVER_IP:$APP_ROOT/public/static/"

echo "🔨  Go Docker imajı build ediliyor..."
$SSH "$SERVER_USER@$SERVER_IP" "cd '$APP_ROOT/go-app' && echo '$SERVER_PASS' | sudo -S docker compose build"

echo "▶️   Container'lar başlatılıyor..."
$SSH "$SERVER_USER@$SERVER_IP" "cd '$APP_ROOT/go-app' && echo '$SERVER_PASS' | sudo -S docker compose up -d --remove-orphans && echo '$SERVER_PASS' | sudo -S docker image prune -f"

echo ""
echo "🎉  Go deploy tamamlandı! → http://$SERVER_IP:8080"
