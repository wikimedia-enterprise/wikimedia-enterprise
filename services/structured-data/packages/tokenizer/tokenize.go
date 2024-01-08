// Package tokenizer converts a string of wikitext into an array of word tokens.
package tokenizer

import (
	"fmt"
	"regexp"
	"strings"
)

// Unified Ideographs, Unified Ideographs Ext A,
// Unified Ideographs Ext. B, Compatibility Ideographs,
// Compatibility Ideographs Suppl.
var chineseMiscChar = []string{
	`\x{4E00}-\x{62FF}`,
	`\x{3400}-\x{4DFF}`,
	`\x{00020000}-\x{000215FF}`,
	`\x{F900}-\x{FAFF}`,
	`\x{0002F800}-\x{0002FA1F}`,
	`\x{6300}-\x{77FF}`,
	`\x{7800}-\x{8CFF}`,
	`\x{8D00}-\x{9FCC}`,
	`\x{00021600}-\x{000230FF}`,
	`\x{00023100}-\x{000245FF}`,
	`\x{00024600}-\x{000260FF}`,
	`\x{00026100}-\x{000275FF}`,
	`\x{00027600}-\x{000290FF}`,
	`\x{00029100}-\x{0002A6DF}`,
	`\x{4E00}-\x{9FCB}`,
	`\x{F900}-\x{FA6A}`,
	`\x{3220}-\x{3243}`,
	`\x{3280}-\x{337F}`,
}

// Hiragana, Katakana, Kanji, Kanji radicals,
// Katakana and Punctuation (Half Width), Misc Symbols and Chars.
var japaneseChar = []string{
	`\x{3041}-\x{3096}`,
	`\x{30A0}-\x{30FF}`,
	`\x{3400}-\x{4DB5}`,
	`\x{2E80}-\x{2FD5}`,
	`\x{FF5F}-\x{FF9F}`,
	`\x{31F0}-\x{31FF}`,
}

// Hangul, Hangul syllables, Hangul Jamo chars.
var koreanChar = []string{
	`\x{AC00}-\x{D7AF}`,
	`\x{1100}-\x{11FF}`,
	`\x{3130}-\x{318F}`,
	`\x{A960}-\x{A97F}`,
	`\x{D7B0}-\x{D7FF}`,
}

// Cyrillic characters.
var cyrillicChar = []string{
	`\x{0400}-\x{04FF}`,
	`\x{0500}-\x{052F}`,
	`\x{2DE0}-\x{2DFF}`,
	`\x{A640}-\x{A69F}`,
	`\x{1C80}-\x{1C8F}`,
	`\x{1D2B}-\x{1D78}`,
	`\x{FE2E}-\x{FE2F}`,
}

// Korean, Japanese and Chinese chars.
var cjkChar = fmt.Sprintf(`%s%s%s`, strings.Join(chineseMiscChar, ""), strings.Join(japaneseChar, ""), strings.Join(koreanChar, ""))

// Regular expresion which matches any number of Korean, Japanese and Chinese char repetitions.
var cjkWord = fmt.Sprintf(`[%s]*(?:[\'’](?:[%s]+))*`, cjkChar, cjkChar)

// Devangari chars.
var devangariChar = `\x{0901}-\x{0963}`

// Arabic chars.
var arabicChar = `\x{0601}-\x{061A}\x{061C}-\x{0669}\x{06D5}-\x{06EF}`

// Bengali chars.
var bengaliChar = `\x{0980}-\x{09FF}`

// Devangari, Arabic and Bengali chars.
var combinedChar = fmt.Sprintf(`%s%s%s%s`, devangariChar, arabicChar, bengaliChar, strings.Join(cyrillicChar, ""))

// Regular expresion which matches any number of Devangari, Arabic and Bengali char repetitions.
var word = fmt.Sprintf(`(?:[^\W]|[%s])[\w%s]*(?:[\'’](?:[\w%s]+))*`, combinedChar, combinedChar, combinedChar)

// URL address starting from slashes until the end.
var urlAddress = `[^\s/$.?#][^|<>\s{}]*`

// Set of protocols.
var plainProtocol = []string{
	"bitcoin", "geo", "magnet", "mailto", "news", "sips?",
	"tel", "urn",
}

// Set of protocols.
var slashedProtocol = []string{"https?", "ftp", "ftps", "git", "gophe", "ircs?",
	"mms", "nntp", "redis", "sftp", "ssh", "svn", "telnet",
	"worldwind", "xmpp",
}

// Regular expresion which matches URLs.
var url = fmt.Sprintf(`(?:(?:%s)\:|(?:(?:%s)\:)?\/\/)%s`, strings.Join(plainProtocol, "|"), strings.Join(slashedProtocol, "|"), urlAddress)

// Regular expresion which matches numbers.
var number = `\b([0-9]+([\,\.][0-9]+)*(e[0-9]+)*|([\.\,][0-9]+)+(e[0-9]+)*)\b`

// Struct which contains match group name and corresponding regular expression.
type matcher struct {
	name  string
	regex string
}

var lexicon = []matcher{
	{
		"break",
		`(?:\n\r?|\r\n)\s*(?:\n\r?|\r\n)+`,
	},
	{
		"whitespace",
		`(?:\n\r?|[^\S\n\r]+)`,
	},
	{
		"tag",
		`</?([a-z][a-z0-9]*)\b[^>]*>`,
	},
	{
		"ref_open",
		`<ref\b(?:/(?:>)|[^>/])*>`,
	},
	{
		"ref_close",
		`</ref\b[^>]*>`,
	},
	{
		"ref_singleton",
		`<ref\b(?:/(?:>)|[^/>])*/>`,
	},
	{
		"cjk_punctuation",
		`[\x{3000}-\x{303F}]`,
	},
	{
		"number",
		number,
	},
	{
		"url",
		url,
	},
	{
		"word",
		word,
	},
	{
		"cjk_word",
		cjkWord,
	},
}

var regex *regexp.Regexp

func init() {
	rstring := []string{}

	for _, m := range lexicon {
		rstring = append(rstring, fmt.Sprintf(`(?P<%s>%s)`, m.name, m.regex))
	}

	regex = regexp.MustCompile(strings.Join(rstring, "|"))
}

// Tokenize accepts a string of wikitext and splits it to word tokens.
// Returns a slice of the word tokens.
func Tokenize(wikitext string) []string {
	matches := regex.FindAllStringSubmatch(strings.ReplaceAll(wikitext, "\\n", " "), -1)
	wordsIndex := regex.SubexpIndex("word")
	cjkWordsIndex := regex.SubexpIndex("cjk_word")
	tokens := []string{}

	for _, match := range matches {
		if token := match[wordsIndex]; len(token) > 0 {
			tokens = append(tokens, token)
		}

		if token := match[cjkWordsIndex]; len(token) > 0 {
			tokens = append(tokens, token)
		}
	}

	return tokens
}
