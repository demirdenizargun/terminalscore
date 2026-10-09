# terminalscore

Terminalde mackolik tarzı futbol takibi: dünyanın her yerinden günün maçları, canlı skor,
maç olayları ve puan durumları. Veri ESPN'in herkese açık API'sinden gelir, API anahtarı gerekmez.

- Maçlar lig büyüklüğüne göre sıralanır (Şampiyonlar Ligi, Premier Lig, LaLiga, ...).
- İlk açılışta favori takımın sorulur; onun maçının olduğu turnuva ve kendi ligi en üstte görünür.
- Puan durumu sekmesinde 46 lig ve kulüp turnuvası arasında gezilebilir.

## Derleme

```sh
go build -o terminalscore .
./terminalscore
```

Windows'ta Linux için derlemek (PowerShell):

```powershell
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o terminalscore-linux-amd64 .
```

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
