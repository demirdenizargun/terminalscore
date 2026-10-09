package espn

// ESPN takım adlarını Türkçe karakter olmadan veriyor; bilinenleri düzeltiyoruz.
var turkishNames = map[string]string{
	"Besiktas":            "Beşiktaş",
	"Fenerbahce":          "Fenerbahçe",
	"Kasimpasa":           "Kasımpaşa",
	"Caykur Rizespor":     "Çaykur Rizespor",
	"Istanbul Basaksehir": "Başakşehir",
	"Istanbul BB":         "Başakşehir",
	"Basaksehir":          "Başakşehir",
	"Genclerbirligi":      "Gençlerbirliği",
	"Genclerbirli":        "Gençlerbirliği",
	"Goztepe":             "Göztepe",
	"Eyupspor":            "Eyüpspor",
	"Corum FK":            "Çorum FK",
	"Fatih Karagumruk":    "Fatih Karagümrük",
	"Karagumruk":          "Karagümrük",
	"Istanbulspor":        "İstanbulspor",
	"Ankaragucu":          "Ankaragücü",
	"MKE Ankaragucu":      "Ankaragücü",
}

func TurkishName(name string) string {
	if tr, ok := turkishNames[name]; ok {
		return tr
	}
	return name
}
