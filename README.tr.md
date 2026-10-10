# terminalscore

[English](README.md)

Terminalde mackolik tarzı canlı skor. Günün bütün maçları, maç olayları ve puan durumları;
veri ESPN'den geliyor, API anahtarı gerekmiyor.

İlk açılışta favori takımını soruyor; o takımın maçı olan turnuva ve kendi ligi listenin
en üstüne çıkıyor. Diğer ligler büyüklük sırasıyla geliyor.

## Kurulum

Linux / macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.sh | sh
```

`~/.local/bin` altına kurar, sudo istemez.

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.ps1 | iex
```

Debian/Ubuntu ve Fedora için `.deb` / `.rpm` paketleri, diğer sistemler için arşivler
[releases](https://github.com/demirdenizargun/terminalscore/releases/latest) sayfasında:

```sh
sudo apt install ./terminalscore_*_linux_amd64.deb
sudo dnf install ./terminalscore_*_linux_amd64.rpm
```

Go ile:

```sh
go install github.com/demirdenizargun/terminalscore@latest
```

## Kullanım

```
terminalscore
```

| Tuş | |
|---|---|
| ← → | gün değiştir (puan durumunda lig) |
| t | bugün |
| ↑ ↓ | maç seç / kaydır |
| PgUp PgDn | 10 maç atla |
| enter | maç detayı |
| esc | geri |
| tab | maçlar / puan durumu |
| f | favori takım |
| r | yenile |
| q | çıkış |

Canlı maç varken liste 10 saniyede, açık maç detayı 5 saniyede bir yenileniyor.
Gol olan maç bir dakika boyunca GOL! diye yanıp sönüyor.

Arayüz dili sistem diline göre Türkçe ya da İngilizce açılıyor. Değiştirmek için
`terminalscore --lang en` ya da ayar dosyasına `"language": "tr"`:
`~/.config/terminalscore/config.json` (Windows'ta `%AppData%\terminalscore\config.json`).
Dosyayı silersen favori takım tekrar sorulur.

## Kaynaktan derleme

```sh
go build -o terminalscore .
```

## Katkı

Çeviri katkılarına açık: `internal/ui/text.go`'ya yeni bir `texts` değeri, `internal/espn/leagues.go`'ya
da lig adlarını eklemek yeterli.

## Lisans

MIT. Maç verileri ESPN'e aittir; bu projenin ESPN ile bir bağlantısı yok. Kullanılan
adresler resmi olarak belgelenmiş bir API değil, her an değişebilir.
