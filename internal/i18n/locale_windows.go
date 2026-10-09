package i18n

import "syscall"

// Windows'ta LANG genelde boş. Bölge biçimine değil görüntüleme diline
// bakıyoruz: bölge en-US, arayüz Türkçe olabiliyor.
func systemLocale() string {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GetUserDefaultUILanguage")
	if proc.Find() != nil {
		return ""
	}
	langID, _, _ := proc.Call()
	if langID&0x3ff == 0x1f { // LANG_TURKISH
		return "tr"
	}
	return ""
}
