package layout

// RU maps Russian (ЙЦУКЕН) characters to US-QWERTY equivalents by physical key
// position. Both lowercase and uppercase are included.

func init() {
	Register(russianMapper())
}

func russianMapper() Mapper {
	m := Mapper{
		// lower
		'ё': "`", 'й': "q", 'ц': "w", 'у': "e", 'к': "r", 'е': "t", 'н': "y", 'г': "u", 'ш': "i", 'щ': "o", 'з': "p", 'х': "[", 'ъ': "]",
		'ф': "a", 'ы': "s", 'в': "d", 'а': "f", 'п': "g", 'р': "h", 'о': "j", 'л': "k", 'д': "l", 'ж': ";", 'э': "'",
		'я': "z", 'ч': "x", 'с': "c", 'м': "v", 'и': "b", 'т': "n", 'ь': "m", 'б': ",", 'ю': ".",
		// upper
		'Ё': "~", 'Й': "Q", 'Ц': "W", 'У': "E", 'К': "R", 'Е': "T", 'Н': "Y", 'Г': "U", 'Ш': "I", 'Щ': "O", 'З': "P", 'Х': "{", 'Ъ': "}",
		'Ф': "A", 'Ы': "S", 'В': "D", 'А': "F", 'П': "G", 'Р': "H", 'О': "J", 'Л': "K", 'Д': "L", 'Ж': ":", 'Э': "\"",
		'Я': "Z", 'Ч': "X", 'С': "C", 'М': "V", 'И': "B", 'Т': "N", 'Ь': "M", 'Б': "<", 'Ю': ">",
	}
	return m
}
