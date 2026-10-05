package reactions

import "strings"

// Phrase holds the two message templates for a /reaction category: one for
// when a target member was mentioned ("{sender} ... {target}"), and one
// for when it wasn't (a self-directed variant), matching the original
// bot's `pharases_list`.
type Phrase struct {
	WithTarget string
	SelfOnly   string
}

// Format renders the phrase for the given sender display name, and
// optionally a target display name (the self-directed variant is used when
// target is empty).
func (p Phrase) Format(sender, target string) string {
	template := p.SelfOnly
	if target != "" {
		template = p.WithTarget
	}
	replacer := strings.NewReplacer("{sender}", sender, "{target}", target)
	return replacer.Replace(template)
}

// Categories lists every /reaction category, in the original bot's order.
// It doubles as the fixed choice list for the /reaction command's "type"
// option -- Discord allows up to 25 choices per option, and there are
// exactly 25 categories here, so no autocomplete is needed.
var Categories = []string{
	"bully", "cuddle", "cry", "hug", "kiss", "lick", "pat", "smug", "bonk",
	"yeet", "blush", "smile", "wave", "highfive", "handhold", "nom", "bite",
	"slap", "kill", "kick", "happy", "wink", "poke", "dance", "cringe",
}

// Phrases maps each category to its message templates, ported verbatim
// from the original bot's pharases_list.
var Phrases = map[string]Phrase{
	"bully":    {"{sender} доебался до {target}", "{sender} так и не понял до кого хотел доебаться и доебался до самого себя"},
	"cuddle":   {"{sender} прижимает к себе {target}", "{sender} хотел обнять кого-то, а обнял аниме девку"},
	"cry":      {"{target} довел до слез {sender}", "Вы довели до слез {sender}, зачем вы так?"},
	"hug":      {"{sender} обнимает {target}", "{sender} обнимает сам себя"},
	"kiss":     {"{sender} целует {target}", "{sender} засосал аниме девочку"},
	"lick":     {"{sender} облизал {target}", "{sender} облизал аниме девочку"},
	"pat":      {"{sender} погладил по головке {target}", "{sender} погладил по головке аниме девочку"},
	"smug":     {"{sender} показывает свое превосходство над {target}", "{sender} показывает свое превосходство"},
	"bonk":     {"{sender} бьет {target}", "{sender} бьет кого-то и промахивается"},
	"yeet":     {"{sender} уебал {target}", "{sender} разъебал всех"},
	"blush":    {"{sender} покраснел от {target}", "{sender} смущается"},
	"smile":    {"{sender} улыбается {target}", "{sender} улыбается"},
	"wave":     {"{sender} помахал {target}", "{sender} машет"},
	"highfive": {"{sender} дает пятюню {target}", "{sender}, ты знаешь что это парная эмоция, да?"},
	"handhold": {"{sender} взял за руку {target}", "{sender} взял за руку своего воображаемого трапика"},
	"nom":      {"{sender} кушает вместе с {target}", "{sender} жрет"},
	"bite":     {"{sender} кусает {target}", "{sender} делает кусь"},
	"slap":     {"{sender} отвесил смачного леща {target}", "{sender} дал аплеуху сам себе, лол"},
	"kill":     {"{sender} совершает уголовно-наказуемое деяние в отношении {target}", "{sender} совершает Роскомнадзор"},
	"kick":     {"{sender} уебал с ноги {target}", "{sender} жостка крутанул вертушку"},
	"happy":    {"{sender} счаслив вместе с {target}", "{sender} радуется"},
	"wink":     {"{sender} подмигнул {target}", "{sender} моргнул одним глазом"},
	"poke":     {"{sender} ткул {target}", "{sender} пытаеться куда ткнуть но не может попасть"},
	"dance":    {"{sender} флексит с {target}", "{sender} отжигает на танцполе"},
	"cringe":   {"{sender} кринжует от {target}", "{sender} на кринже ваще"},
}
