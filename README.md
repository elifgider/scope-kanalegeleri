<div align="center">
  <img src="https://i.hizliresim.com/n94gt6w.png" alt="Scope Logo" width="250" />
  <h1>Scope Kanal Eğeleri</h1>
  <p>Modern, Hızlı ve Güvenilir Endodontik Sipariş Yönetim Sistemi</p>

  <a href="https://kanalegeleri.com.tr/"><strong>Canlı Siteyi Görüntüle →</strong></a>

  <br />
  <br />

  ![Hero Preview](https://i.hizliresim.com/jlyg6uj.png)
</div>

<hr />

### 🛠️ Teknoloji Yığını

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white" />
  <img src="https://img.shields.io/badge/Gin-008080?style=for-the-badge&logo=gin&logoColor=white" />
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" />
  <img src="https://img.shields.io/badge/Bootstrap-7952B3?style=for-the-badge&logo=bootstrap&logoColor=white" />
</p>

---

### 🏗️ Proje Mimarisi (Architecture Overview)

Proje, sürdürülebilir ve test edilebilir bir yapı için **Layered Architecture (Katmanlı Mimari)** prensiplerine göre yapılandırılmıştır:

- **Handler Layer:** HTTP isteklerini karşılar, input validasyonu yapar ve sonuçları döner.
- **Service Layer:** İş mantığının (Business Logic) merkezi katmanıdır.
- **Repository Layer:** Veritabanı (PostgreSQL) işlemlerini soyutlar.
- **Domain/Models:** Veri yapılarını ve ortak interface tanımlarını içerir.

### 🌐 API Tasarımı

Uygulama, frontend ile JSON tabanlı bir API üzerinden haberleşir. Önemli uç noktalar:

| Method | Endpoint | Açıklama |
| :--- | :--- | :--- |
| `GET` | `/api/products` | Tüm ürünleri filtrelerle birlikte döner. |
| `POST` | `/api/orders` | Yeni bir sipariş talebi oluşturur ve bildirim tetikler. |
| `POST` | `/admin/login` | Yönetici paneli için güvenli oturum açma işlemi. |

### 💾 Veritabanı Tasarımı

PostgreSQL üzerinde **GORM** kullanılarak modellenen yapı, ilişkisel veri bütünlüğünü sağlar:
- `products`: Ürün adı, kategori, açıklama ve görsel URL bilgilerini tutar.
- `orders`: Müşteri bilgileri ve sipariş notlarını saklar.
- `order_items`: Siparişle ilişkili ürünlerin miktar bilgilerini tutar.

### ❓ Neden Go ve Gin?

- **Performans:** Go'nun düşük bellek kullanımı ve yüksek hızı, uygulamanın anlık yüklere hızlı yanıt vermesini sağlar.
- **Gin Web Framework:** Minimalist yapısı ve gelişmiş middleware desteği ile hızlı geliştirme imkanı sunar.
- **Concurrency:** Siparişlerin kaydedilmesi ve bildirimlerin gönderilmesi gibi işlemler Go rutinleri ile verimli yönetilir.

### 💡 Zorluklar ve Çözümler (Challenges & Solutions)

- **Challenge:** Veritabanı geçişleri ve şema yönetimi.
- **Solution:** GORM'un `AutoMigrate` özelliği ile veritabanı şeması kod ile senkronize tutuldu, manuel SQL yönetiminden kaçınıldı.
- **Challenge:** Çoklu sipariş kalemlerinin yönetimi.
- **Solution:** Sipariş ve ürünler arasında bire-çok ilişki kurularak, JSON tabanlı dinamik sipariş işleme mantığı geliştirildi.

---

### 📸 Arayüzden Kareler

<div align="center">
  <img src="https://i.hizliresim.com/8l7yl46.png" alt="Ürünler" width="800" />
</div>

---

### 🚀 Hızlı Başlangıç

1.  `.env` dosyasını yapılandırın.
2.  Konteynerleri kaldırın:
    ```bash
    docker-compose up -d --build
    ```

<p align="center">© 2026 Scope Endo - Tüm Hakları Saklıdır.</p>
