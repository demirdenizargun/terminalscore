# terminalscore

Terminalde mackolik tarzı futbol takibi: dünyanın her yerinden günün maçları, canlı skor,
maç olayları ve puan durumları. Veri ESPN'in herkese açık API'sinden gelir, API anahtarı gerekmez.

- Maçlar lig büyüklüğüne göre sıralanır (Şampiyonlar Ligi, Premier Lig, LaLiga, ...).
- İlk açılışta favori takımın sorulur; onun maçının olduğu turnuva ve kendi ligi en üstte görünür.
- Puan durumu sekmesinde 46 lig ve kulüp turnuvası arasında gezilebilir.

## Kurulum

Kurulumdan sonra terminale `terminalscore` yazman yeterli. Go ya da başka bir şey kurmana gerek yok.

**Linux / macOS**

```sh
curl -fsSL https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.sh | sh
```

Program `~/.local/bin` klasörüne kurulur (yönetici izni gerekmez). Betik, indirdiği dosyanın
bozulmadığını sağlama değeriyle doğrular.

**Ubuntu / Debian** — [sürüm sayfasından](https://github.com/demirdenizargun/terminalscore/releases/latest) `.deb` dosyasını indir:

```sh
sudo apt install ./terminalscore_*_linux_amd64.deb
```

**Fedora / openSUSE** — sürüm sayfasından `.rpm` dosyasını indir:

```sh
sudo dnf install ./terminalscore_*_linux_amd64.rpm
```

**Windows** (PowerShell)

```powershell
irm https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.ps1 | iex
```

**Elle kurulum** — [sürüm sayfasından](https://github.com/demirdenizargun/terminalscore/releases/latest)
sistemine uygun arşivi indir, içindeki `terminalscore` dosyasını PATH'teki bir klasöre koy.
ARM işlemciler (Raspberry Pi, Apple M serisi) için `arm64` dosyalarını seç.

**Go kullananlar için**

```sh
go install github.com/demirdenizargun/terminalscore@latest
```

Kurulu sürümü görmek için: `terminalscore --version`

## Geliştirme

```sh
go build -o terminalscore .
./terminalscore
```

Windows'ta Linux için derlemek (PowerShell):

```powershell
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o terminalscore-linux-amd64 .
```

### Sürüm yayınlama

Sürümler [GoReleaser](https://goreleaser.com) ile GitHub Actions üzerinde otomatik hazırlanır.
Bir etiket göndermek yeterli:

```sh
git tag v0.1.0
git push origin v0.1.0
```

Bütün platformların dosyaları, `.deb`/`.rpm` paketleri ve `checksums.txt` sürüm sayfasına yüklenir.
Yayınlamadan önce yerelde denemek için: `goreleaser release --snapshot --clean` (çıktı `dist/` klasörüne).

## Tuşlar

| Tuş | İşlev |
|---|---|
| ← / → | Önceki / sonraki gün (puan durumunda: önceki / sonraki lig) |
| t | Bugüne dön |
| ↑ / ↓ | Maç seç / kaydır |
| PgUp / PgDn | 10 maç yukarı / aşağı |
| enter | Maç detayı |
| esc | Geri |
| tab | Maçlar / Puan Durumu |
| f | Favori takımı değiştir |
| r | Yenile |
| q | Çıkış |

Canlı maç varken maç listesi 10 saniyede, açık maç detayı 5 saniyede bir yenilenir
(ESPN verisi sunucuda bu sürelerde güncelleniyor). Gol olan maç bir dakika boyunca
yanıp sönen **GOL!** etiketiyle işaretlenir.

## Ayarlar

Favori takım şu dosyada saklanır; silersen ilk açılıştaki soru tekrar gelir:

- Linux: `~/.config/terminalscore/config.json`
- Windows: `%AppData%\terminalscore\config.json`

## Lisans ve veri

Kod [MIT lisansı](LICENSE) ile dağıtılır.

Maç verileri ESPN'e aittir. Bu proje ESPN ile bağlantılı değildir ve ESPN tarafından
desteklenmez; kullanılan adresler ESPN'in resmi olarak belgelenmiş bir API'si değildir
ve haber verilmeden değişebilir.
