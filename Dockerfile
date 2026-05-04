# En küçük ve hızlı imaj
FROM debian:bookworm-slim

# Gerekli sistem paketleri (SSL ve saat dilimi için)
RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Bilgisayarımızda derlediğimiz binary dosyasını kopyalıyoruz
COPY app_binary .

# Diğer statik klasörleri kopyalıyoruz
COPY templates/ ./templates/
COPY public/ ./public/
COPY .env .

# Klasör izinleri
RUN mkdir -p uploads data && chmod 777 uploads data

EXPOSE 8080

# Uygulamayı başlat
CMD ["./app_binary"]
