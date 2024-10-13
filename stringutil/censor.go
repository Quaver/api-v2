package stringutil

import (
	"slices"
	"strings"
)

var CensoredClanTags = []string{
	"anal",
	"anus",
	"arab",
	"arse",
	"ass",
	"a55",
	"boob",
	"b00b",
	"butt",
	"chav",
	"clit",
	"cock",
	"c0ck",
	"coon",
	"c00n",
	"crap",
	"cum",
	"cumm",
	"cunt",
	"damn",
	"dick",
	"dyke",
	"fag",
	"fags",
	"fuck",
	"fuk",
	"gay",
	"gays",
	"gook",
	"hell",
	"homo",
	"jiz ",
	"jizz",
	"kink",
	"kkk",
	"kum",
	"milf",
	"nazi",
	"nig",
	"nigs",
	"niga",
	"nigr",
	"oral",
	"orgy",
	"phuk",
	"phuq",
	"pi55",
	"piss",
	"porn",
	"pusy",
	"rape",
	"scat",
	"sex",
	"s3x",
	"shag",
	"shat",
	"shit",
	"slut",
	"smut",
	"spic",
	"spik",
	"tard",
	"tit",
	"t1t",
	"tits",
	"t1ts",
	"t1t5",
	"twat",
	"wank",
}

// IsClanTagCensored Returns if the provided clan tag is censored
func IsClanTagCensored(tag string) bool {
	lowerTag := strings.ToLower(tag)

	return slices.ContainsFunc(CensoredClanTags, func(s string) bool {
		return strings.ToLower(s) == lowerTag
	})
}
