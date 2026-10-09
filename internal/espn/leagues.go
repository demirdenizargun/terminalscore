package espn

type LeagueKind int

const (
	Domestic    LeagueKind = iota // puan durumu var, takımlar favori aramasında
	Continental                   // ŞL, Avrupa Ligi vb.
	Cup
	National // milli takımlar
)

type League struct {
	Slug string // ör. "tur.1"
	ID   string // uid'deki "l:3946"
	Name string
	Kind LeagueKind
}

func (l League) HasStandings() bool { return l.Kind == Domestic || l.Kind == Continental }

// Büyükten küçüğe; maç listesindeki lig sırası buradan geliyor.
var Leagues = []League{
	{"fifa.world", "606", "Dünya Kupası", National},
	{"uefa.euro", "781", "Avrupa Şampiyonası", National},
	{"uefa.champions", "775", "UEFA Şampiyonlar Ligi", Continental},
	{"eng.1", "700", "İngiltere Premier Lig", Domestic},
	{"esp.1", "740", "İspanya LaLiga", Domestic},
	{"ita.1", "730", "İtalya Serie A", Domestic},
	{"ger.1", "720", "Almanya Bundesliga", Domestic},
	{"fra.1", "710", "Fransa Ligue 1", Domestic},
	{"uefa.europa", "776", "UEFA Avrupa Ligi", Continental},
	{"uefa.nations", "2395", "UEFA Uluslar Ligi", National},
	{"fifa.worldq.uefa", "786", "Dünya Kupası Elemeleri (Avrupa)", National},
	{"uefa.europa.conf", "20296", "UEFA Konferans Ligi", Continental},
	{"fifa.cwc", "5501", "FIFA Kulüpler Dünya Kupası", Continental},
	{"uefa.super_cup", "5462", "UEFA Süper Kupa", Cup},
	{"por.1", "715", "Portekiz Liga Portugal", Domestic},
	{"ned.1", "725", "Hollanda Eredivisie", Domestic},
	{"tur.1", "3946", "Trendyol Süper Lig", Domestic},
	{"bel.1", "3901", "Belçika Pro League", Domestic},
	{"sco.1", "735", "İskoçya Premiership", Domestic},
	{"bra.1", "630", "Brezilya Serie A", Domestic},
	{"arg.1", "745", "Arjantin Liga Profesional", Domestic},
	{"conmebol.libertadores", "783", "Copa Libertadores", Continental},
	{"usa.1", "770", "ABD MLS", Domestic},
	{"ksa.1", "21231", "Suudi Arabistan Pro Lig", Domestic},
	{"mex.1", "760", "Meksika Liga MX", Domestic},
	{"eng.2", "3914", "İngiltere Championship", Domestic},
	{"esp.2", "3921", "İspanya LaLiga 2", Domestic},
	{"ita.2", "3931", "İtalya Serie B", Domestic},
	{"ger.2", "3927", "Almanya 2. Bundesliga", Domestic},
	{"fra.2", "3926", "Fransa Ligue 2", Domestic},
	{"eng.fa", "3918", "İngiltere FA Cup", Cup},
	{"esp.copa_del_rey", "3951", "İspanya Kral Kupası", Cup},
	{"ita.coppa_italia", "3956", "İtalya Kupası", Cup},
	{"ger.dfb_pokal", "3954", "Almanya Kupası", Cup},
	{"fra.coupe_de_france", "3952", "Fransa Kupası", Cup},
	{"eng.league_cup", "3920", "İngiltere Lig Kupası", Cup},
	{"aut.1", "3907", "Avusturya Bundesliga", Domestic},
	{"sui.1", "3944", "İsviçre Super League", Domestic},
	{"gre.1", "3955", "Yunanistan Super League", Domestic},
	{"den.1", "3913", "Danimarka Superliga", Domestic},
	{"nor.1", "3960", "Norveç Eliteserien", Domestic},
	{"swe.1", "3945", "İsveç Allsvenskan", Domestic},
	{"cze.1", "5347", "Çekya 1. Lig", Domestic},
	{"rus.1", "3939", "Rusya Premier Lig", Domestic},
	{"rou.1", "10747", "Romanya Liga 1", Domestic},
	{"isr.1", "10750", "İsrail Premier Lig", Domestic},
	{"cyp.1", "5346", "Kıbrıs 1. Lig", Domestic},
	{"ned.2", "3933", "Hollanda 2. Lig", Domestic},
	{"eng.3", "3915", "İngiltere League One", Domestic},
	{"jpn.1", "750", "Japonya J1 League", Domestic},
	{"chn.1", "8376", "Çin Süper Lig", Domestic},
	{"aus.1", "3906", "Avustralya A-League", Domestic},
	{"col.1", "650", "Kolombiya Primera A", Domestic},
	{"chi.1", "640", "Şili Primera División", Domestic},
	{"uru.1", "680", "Uruguay Primera División", Domestic},
	{"rsa.1", "3937", "Güney Afrika Premiership", Domestic},
	{"ind.1", "8316", "Hindistan Süper Lig", Domestic},
	{"fifa.friendly", "3922", "Milli Hazırlık Maçları", National},
	{"club.friendly", "19834", "Kulüp Hazırlık Maçları", Cup},
}

var leagueIndex = func() map[string]int {
	idx := make(map[string]int, len(Leagues))
	for i, l := range Leagues {
		idx[l.ID] = i
		idx[l.Slug] = i
	}
	return idx
}()

// Listede olmayan lig en sona düşer.
func LeagueRank(idOrSlug string) int {
	if i, ok := leagueIndex[idOrSlug]; ok {
		return i
	}
	return len(Leagues)
}

func FindLeague(idOrSlug string) (League, bool) {
	if i, ok := leagueIndex[idOrSlug]; ok {
		return Leagues[i], true
	}
	return League{}, false
}
