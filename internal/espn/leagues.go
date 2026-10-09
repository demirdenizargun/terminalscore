package espn

type LeagueKind int

const (
	Domestic    LeagueKind = iota // puan durumu var, takımlar favori aramasında
	Continental                   // ŞL, Avrupa Ligi vb.
	Cup
	National // milli takımlar
)

type League struct {
	Slug   string // ör. "tur.1"
	ID     string // uid'deki "l:3946"
	NameTR string
	NameEN string
	Kind   LeagueKind
}

func (l League) HasStandings() bool { return l.Kind == Domestic || l.Kind == Continental }

// Büyükten küçüğe; maç listesindeki lig sırası buradan geliyor.
var Leagues = []League{
	{"fifa.world", "606", "Dünya Kupası", "FIFA World Cup", National},
	{"uefa.euro", "781", "Avrupa Şampiyonası", "UEFA Euro", National},
	{"uefa.champions", "775", "UEFA Şampiyonlar Ligi", "UEFA Champions League", Continental},
	{"eng.1", "700", "İngiltere Premier Lig", "Premier League", Domestic},
	{"esp.1", "740", "İspanya LaLiga", "LaLiga", Domestic},
	{"ita.1", "730", "İtalya Serie A", "Serie A", Domestic},
	{"ger.1", "720", "Almanya Bundesliga", "Bundesliga", Domestic},
	{"fra.1", "710", "Fransa Ligue 1", "Ligue 1", Domestic},
	{"uefa.europa", "776", "UEFA Avrupa Ligi", "UEFA Europa League", Continental},
	{"uefa.nations", "2395", "UEFA Uluslar Ligi", "UEFA Nations League", National},
	{"fifa.worldq.uefa", "786", "Dünya Kupası Elemeleri (Avrupa)", "World Cup Qualifying (UEFA)", National},
	{"uefa.europa.conf", "20296", "UEFA Konferans Ligi", "UEFA Conference League", Continental},
	{"fifa.cwc", "5501", "FIFA Kulüpler Dünya Kupası", "FIFA Club World Cup", Continental},
	{"uefa.super_cup", "5462", "UEFA Süper Kupa", "UEFA Super Cup", Cup},
	{"por.1", "715", "Portekiz Liga Portugal", "Liga Portugal", Domestic},
	{"ned.1", "725", "Hollanda Eredivisie", "Eredivisie", Domestic},
	{"tur.1", "3946", "Trendyol Süper Lig", "Süper Lig", Domestic},
	{"bel.1", "3901", "Belçika Pro League", "Belgian Pro League", Domestic},
	{"sco.1", "735", "İskoçya Premiership", "Scottish Premiership", Domestic},
	{"bra.1", "630", "Brezilya Serie A", "Brasileirão Série A", Domestic},
	{"arg.1", "745", "Arjantin Liga Profesional", "Argentine Liga Profesional", Domestic},
	{"conmebol.libertadores", "783", "Copa Libertadores", "Copa Libertadores", Continental},
	{"usa.1", "770", "ABD MLS", "MLS", Domestic},
	{"ksa.1", "21231", "Suudi Arabistan Pro Lig", "Saudi Pro League", Domestic},
	{"mex.1", "760", "Meksika Liga MX", "Liga MX", Domestic},
	{"eng.2", "3914", "İngiltere Championship", "EFL Championship", Domestic},
	{"esp.2", "3921", "İspanya LaLiga 2", "LaLiga 2", Domestic},
	{"ita.2", "3931", "İtalya Serie B", "Serie B", Domestic},
	{"ger.2", "3927", "Almanya 2. Bundesliga", "2. Bundesliga", Domestic},
	{"fra.2", "3926", "Fransa Ligue 2", "Ligue 2", Domestic},
	{"eng.fa", "3918", "İngiltere FA Cup", "FA Cup", Cup},
	{"esp.copa_del_rey", "3951", "İspanya Kral Kupası", "Copa del Rey", Cup},
	{"ita.coppa_italia", "3956", "İtalya Kupası", "Coppa Italia", Cup},
	{"ger.dfb_pokal", "3954", "Almanya Kupası", "DFB-Pokal", Cup},
	{"fra.coupe_de_france", "3952", "Fransa Kupası", "Coupe de France", Cup},
	{"eng.league_cup", "3920", "İngiltere Lig Kupası", "EFL Cup", Cup},
	{"aut.1", "3907", "Avusturya Bundesliga", "Austrian Bundesliga", Domestic},
	{"sui.1", "3944", "İsviçre Super League", "Swiss Super League", Domestic},
	{"gre.1", "3955", "Yunanistan Super League", "Greek Super League", Domestic},
	{"den.1", "3913", "Danimarka Superliga", "Danish Superliga", Domestic},
	{"nor.1", "3960", "Norveç Eliteserien", "Eliteserien", Domestic},
	{"swe.1", "3945", "İsveç Allsvenskan", "Allsvenskan", Domestic},
	{"cze.1", "5347", "Çekya 1. Lig", "Czech First League", Domestic},
	{"rus.1", "3939", "Rusya Premier Lig", "Russian Premier League", Domestic},
	{"rou.1", "10747", "Romanya Liga 1", "Romanian Liga I", Domestic},
	{"isr.1", "10750", "İsrail Premier Lig", "Israeli Premier League", Domestic},
	{"cyp.1", "5346", "Kıbrıs 1. Lig", "Cypriot First Division", Domestic},
	{"ned.2", "3933", "Hollanda 2. Lig", "Eerste Divisie", Domestic},
	{"eng.3", "3915", "İngiltere League One", "EFL League One", Domestic},
	{"jpn.1", "750", "Japonya J1 League", "J1 League", Domestic},
	{"chn.1", "8376", "Çin Süper Lig", "Chinese Super League", Domestic},
	{"aus.1", "3906", "Avustralya A-League", "A-League Men", Domestic},
	{"col.1", "650", "Kolombiya Primera A", "Colombian Primera A", Domestic},
	{"chi.1", "640", "Şili Primera División", "Chilean Primera División", Domestic},
	{"uru.1", "680", "Uruguay Primera División", "Uruguayan Primera División", Domestic},
	{"rsa.1", "3937", "Güney Afrika Premiership", "South African Premiership", Domestic},
	{"ind.1", "8316", "Hindistan Süper Lig", "Indian Super League", Domestic},
	{"fifa.friendly", "3922", "Milli Hazırlık Maçları", "International Friendlies", National},
	{"club.friendly", "19834", "Kulüp Hazırlık Maçları", "Club Friendlies", Cup},
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
