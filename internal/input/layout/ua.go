package layout

// UA maps Ukrainian (ЙЦУКЕН) characters to US-QWERTY equivalents by physical
// key position. It mirrors the RU map with Ukrainian-specific letters.

func init() { Register(ukrainianMapper()) }

func ukrainianMapper() Mapper {
	m := Mapper{
		// lower (shared with RU-like positions)
		'й': "q", 'ц': "w", 'у': "e", 'к': "r", 'е': "t", 'н': "y", 'г': "u", 'ш': "i", 'щ': "o", 'з': "p", 'х': "[",
		'ф': "a", 'в': "d", 'а': "f", 'п': "g", 'р': "h", 'о': "j", 'л': "k", 'д': "l", 'ж': ";",
		'я': "z", 'ч': "x", 'с': "c", 'м': "v", 'т': "n", 'ь': "m", 'б': ",", 'ю': ".",
		// Ukrainian-specific replacements vs RU
		'ґ': "`", // replaces RU 'ё'
		'і': "s", // replaces RU 'ы'
		'є': "'", // replaces RU 'э'
		'ї': "]", // replaces RU 'ъ'

		// upper
		'Й': "Q", 'Ц': "W", 'У': "E", 'К': "R", 'Е': "T", 'Н': "Y", 'Г': "U", 'Ш': "I", 'Щ': "O", 'З': "P", 'Х': "{",
		'Ф': "A", 'В': "D", 'А': "F", 'П': "G", 'Р': "H", 'О': "J", 'Л': "K", 'Д': "L", 'Ж': ":",
		'Я': "Z", 'Ч': "X", 'С': "C", 'М': "V", 'Т': "N", 'Ь': "M", 'Б': "<", 'Ю': ">",
		'Ґ': "~",
		'І': "S",
		'Є': "\"",
		'Ї': "}",
	}
	return m
}
