package service

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Dict 内存字典：中文 → 各语言译文
type Dict struct {
	Entries map[string]map[string]string
	Misses  map[string]struct{} // 翻译过程中未命中的中文词
}

func (d *Dict) recordMiss(term string) {
	if d == nil {
		return
	}
	term = strings.TrimSpace(term)
	if term == "" || !hasChinese(term) {
		return
	}
	if d.Misses == nil {
		d.Misses = make(map[string]struct{})
	}
	d.Misses[term] = struct{}{}
}

// MissList 返回去重后的未命中词（按字典序）
func (d *Dict) MissList() []string {
	if d == nil || len(d.Misses) == 0 {
		return nil
	}
	list := make([]string, 0, len(d.Misses))
	for w := range d.Misses {
		list = append(list, w)
	}
	sort.Strings(list)
	return list
}

var chineseAliases = map[string][]string{
	"成分": {"成份"},
	"成份": {"成分"},
	"面料": {"面布"},
	"里料": {"里布"},
}

var (
	wholeParenRe  = regexp.MustCompile(`^[（(]([^（）()]+)[）)]$`)
	inlineParenRe = regexp.MustCompile(`[（(]([^（）()]+)[）)]`)
	hasParenRe    = regexp.MustCompile(`[（(][^（）()]+[）)]`)
	pctNameRe     = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*[%％]\s*([\p{Han}][\p{Han}\w]*)`)
	pctOnlyRe     = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*[%％]`)
	splitTextRe   = regexp.MustCompile(`（[^）]*）|[\p{Han}]+|[^\p{Han}]+`)
	hasChineseRe  = regexp.MustCompile(`\p{Han}`)
)

func (d *Dict) lookup(chinese, field string) string {
	trimmed := strings.TrimSpace(chinese)
	if trimmed == "" || d == nil || d.Entries == nil {
		return ""
	}

	keys := []string{
		trimmed,
		"（" + trimmed + "）",
		"(" + trimmed + ")",
	}
	keys = append(keys, chineseAliases[trimmed]...)
	keys = append(keys, chineseAliases["（"+trimmed+"）"]...)

	queryHasParen := hasOuterParen(trimmed)
	seen := map[string]bool{}
	for _, key := range keys {
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		if entry, ok := d.Entries[key]; ok {
			if v := strings.TrimSpace(entry[field]); v != "" {
				return normalizeLookupValue(v, queryHasParen)
			}
		}
		for _, alias := range chineseAliases[key] {
			if entry, ok := d.Entries[alias]; ok {
				if v := strings.TrimSpace(entry[field]); v != "" {
					return normalizeLookupValue(v, queryHasParen)
				}
			}
		}
	}
	return ""
}

func hasOuterParen(s string) bool {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) < 2 {
		return false
	}
	first, last := runes[0], runes[len(runes)-1]
	return (first == '（' || first == '(') && (last == '）' || last == ')')
}

func stripOuterParen(s string) string {
	t := strings.TrimSpace(s)
	for hasOuterParen(t) {
		runes := []rune(t)
		t = strings.TrimSpace(string(runes[1 : len(runes)-1]))
	}
	return t
}

// 查询词本身无括号时，去掉译文外层括号，避免与原文括号叠成 ((xxx))
func normalizeLookupValue(v string, queryHasParen bool) string {
	if queryHasParen {
		return v
	}
	return stripOuterParen(v)
}

// TranslateText 将中文翻译为目标语言（逻辑对齐前端 dictionary.js）
func (d *Dict) TranslateText(chinese, lang string) string {
	if d == nil || chinese == "" {
		return chinese
	}
	trimmed := strings.TrimSpace(chinese)
	if trimmed == "" {
		return ""
	}

	lines := strings.Split(trimmed, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = d.translateLine(line, lang)
	}
	return strings.Join(out, "\n")
}

func (d *Dict) translateLine(chinese, field string) string {
	trimmed := strings.TrimSpace(chinese)
	if trimmed == "" {
		return ""
	}

	if m := wholeParenRe.FindStringSubmatch(trimmed); m != nil {
		inner := strings.TrimSpace(m[1])
		if t := d.lookup(inner, field); t != "" {
			return wrapInParen(t)
		}
		d.recordMiss(inner)
		return "[" + inner + "]"
	}

	if hasParenRe.MatchString(trimmed) {
		var b strings.Builder
		last := 0
		for _, loc := range inlineParenRe.FindAllStringSubmatchIndex(trimmed, -1) {
			before := trimmed[last:loc[0]]
			if before != "" {
				b.WriteString(d.translatePlain(before, field))
			}
			inner := strings.TrimSpace(trimmed[loc[2]:loc[3]])
			if t := d.lookup(inner, field); t != "" {
				b.WriteString(wrapInParen(t))
			} else {
				d.recordMiss(inner)
				b.WriteString("[" + inner + "]")
			}
			last = loc[1]
		}
		tail := trimmed[last:]
		if tail != "" {
			b.WriteString(d.translatePlain(tail, field))
		}
		return b.String()
	}

	return d.translatePlain(trimmed, field)
}

func (d *Dict) translatePlain(chinese, field string) string {
	trimmed := strings.TrimSpace(chinese)
	if trimmed == "" {
		return ""
	}

	label, value, ok := splitByColon(trimmed)
	if ok {
		translatedLabel := d.lookup(label, field)
		displayLabel := label + ":"
		if translatedLabel != "" {
			displayLabel = translatedLabel + ":"
		} else {
			d.recordMiss(label)
		}
		if value == "" {
			return displayLabel
		}
		translatedValue := d.translateCompositionValue(value, field)
		if translatedValue == "" {
			translatedValue = d.translatePlainText(value, field)
		}
		if translatedValue != "" {
			return displayLabel + " " + translatedValue
		}
		return displayLabel
	}

	if exact := d.lookup(trimmed, field); exact != "" {
		return exact
	}
	if comp := d.translateCompositionValue(trimmed, field); comp != "" {
		return comp
	}
	return d.translatePlainText(trimmed, field)
}

func splitByColon(s string) (label, value string, ok bool) {
	idx := -1
	for i, r := range s {
		if r == '：' || r == ':' {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return "", "", false
	}
	label = strings.TrimSpace(s[:idx])
	_, size := utf8.DecodeRuneInString(s[idx:])
	value = strings.TrimSpace(s[idx+size:])
	return label, value, true
}

func (d *Dict) translatePlainText(chinese, field string) string {
	parts := splitTextRe.FindAllString(chinese, -1)
	if parts == nil {
		d.recordMiss(chinese)
		return chinese
	}
	var b strings.Builder
	for _, p := range parts {
		if !hasChineseRe.MatchString(p) {
			b.WriteString(p)
			continue
		}
		term := strings.TrimSpace(p)
		if t := d.lookup(term, field); t != "" {
			b.WriteString(t)
		} else {
			d.recordMiss(term)
			b.WriteString(p)
		}
	}
	return b.String()
}

func (d *Dict) translateCompositionValue(value, field string) string {
	hasMatch := false
	result := pctNameRe.ReplaceAllStringFunc(value, func(match string) string {
		hasMatch = true
		sub := pctNameRe.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		pct, name := sub[1], strings.TrimSpace(sub[2])
		if t := d.lookup(name, field); t != "" && !strings.HasPrefix(t, "[") {
			return pct + "% " + t
		}
		d.recordMiss(name)
		return pct + "%" + name
	})
	if hasMatch {
		return result
	}

	pctOnlyResult := pctOnlyRe.ReplaceAllString(value, "$1%")
	if pctOnlyResult != value {
		return pctOnlyResult
	}
	return ""
}

// HasUntranslated 判断译文是否仍含未翻译痕迹
func HasUntranslated(src, dst string) bool {
	src = strings.TrimSpace(src)
	dst = strings.TrimSpace(dst)
	if src == "" {
		return false
	}
	if dst == "" {
		return true
	}
	if strings.Contains(dst, "[") && strings.Contains(dst, "]") {
		return true
	}
	if hasChinese(src) && src == dst {
		return true
	}
	return false
}

func hasChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// wrapInParen 用半角括号包裹译文；若译文本身已带括号则先剥掉，避免 ((xxx))
func wrapInParen(translated string) string {
	return "(" + stripOuterParen(translated) + ")"
}
