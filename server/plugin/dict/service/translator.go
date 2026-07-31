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
	pctNameRe      = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*[` + pctMarks + `](\s*)([\p{Han}][\p{Han}\w]*)`)
	pctOnlyRe      = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*[` + pctMarks + `]`)
	pctBeforeNumRe = regexp.MustCompile(`[` + pctMarks + `]\s*(\d+(?:\.\d+)?)`) // %84.8 → 84.8%
	numSpacePctRe  = regexp.MustCompile(`(\d+(?:\.\d+)?)\s+[` + pctMarks + `]`) // 84.8 % → 84.8%
	numArabPctRe   = regexp.MustCompile(`(\d+(?:\.\d+)?)[` + pctMarks + `]`)    // 统一百分号为 ASCII %
	numThenPctRe       = regexp.MustCompile(`(\d+(?:\.\d+)?)%`) // 84.8%
	pctStuckArabicRe   = regexp.MustCompile(`%(\p{Arabic})`)    // 84.8%البولي → 84.8% البولي
	splitTextRe        = regexp.MustCompile(`（[^）]*）|[\p{Han}]+|[^\p{Han}]+`)
	hasChineseRe       = regexp.MustCompile(`\p{Han}`)
)

// ASCII % / 全角 ％ / 阿语 ٪
const pctMarks = "%％\u066A"

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

// TranslateText 将中文翻译为目标语言（统一翻译引擎）
// 保留原文换行与行首缩进，使成分等多行结构与中文对齐。
func (d *Dict) TranslateText(chinese, lang string) string {
	if d == nil || chinese == "" {
		return chinese
	}
	if strings.TrimSpace(chinese) == "" {
		return chinese
	}

	lines := strings.Split(chinese, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = d.translateLine(line, lang)
	}
	result := strings.Join(out, "\n")
	result = normalizePercentOrder(result)
	// 阿语 RTL 下仍保持「数字%」逻辑顺序，包成 LTR 单词防显示颠倒
	if lang == "arabic" {
		result = protectArabicPercents(result)
	}
	return result
}

// normalizePercentOrder 统一百分比为「数字%」（百分号在右边）：
// %84.8 / ٪84.8 → 84.8% ；84.8 % → 84.8% ；84.8٪ → 84.8%
func normalizePercentOrder(s string) string {
	if s == "" {
		return s
	}
	if !strings.ContainsAny(s, pctMarks) {
		return s
	}
	s = pctBeforeNumRe.ReplaceAllString(s, "$1%")
	s = numSpacePctRe.ReplaceAllString(s, "$1%")
	s = numArabPctRe.ReplaceAllString(s, "$1%")
	return s
}

// stripBidiControls 去掉方向控制符（LRM/RLM/LRE/RLE/LRO/RLO/FSI/PDI 等）
func stripBidiControls(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\u200E', '\u200F', // LRM RLM
			'\u202A', '\u202B', '\u202C', '\u202D', '\u202E', // LRE RLE PDF LRO RLO
			'\u2066', '\u2067', '\u2068', '\u2069': // LRI RLI FSI PDI
			return -1
		default:
			return r
		}
	}, s)
}

// protectArabicPercents 按行处理：若该行百分比前面已有阿语字母（如 النسيج: 84.8%），
// 逻辑序写成 %84.8，夹在阿语中从左往右看才是 84.8%；纯缩进续行（15.2%）保持 15.2%。
func protectArabicPercents(s string) string {
	if s == "" {
		return s
	}
	s = stripBidiControls(s)
	s = strings.ReplaceAll(s, "\u2060", "")
	s = strings.ReplaceAll(s, "％", "%")
	if !strings.Contains(s, "%") {
		return s
	}
	s = pctStuckArabicRe.ReplaceAllString(s, "% $1")

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = protectArabicPercentLine(line)
	}
	return strings.Join(lines, "\n")
}

func protectArabicPercentLine(line string) string {
	locs := numThenPctRe.FindAllStringSubmatchIndex(line, -1)
	if len(locs) == 0 {
		return line
	}
	var b strings.Builder
	last := 0
	for _, loc := range locs {
		b.WriteString(line[last:loc[0]])
		num := line[loc[2]:loc[3]]
		before := line[:loc[0]]
		if hasArabicLetters(before) {
			// النسيج: 84.8% → النسيج: %84.8 （视觉上从左往右为 84.8%）
			b.WriteString("%")
			b.WriteString(num)
		} else {
			// 纯缩进续行保持 15.2%
			b.WriteString(num)
			b.WriteString("%")
		}
		last = loc[1]
	}
	b.WriteString(line[last:])
	return b.String()
}

func hasArabicLetters(s string) bool {
	for _, r := range s {
		if r >= 0x0600 && r <= 0x06FF { // Arabic
			return true
		}
		if r >= 0x0750 && r <= 0x077F { // Arabic Supplement
			return true
		}
		if r >= 0x08A0 && r <= 0x08FF { // Arabic Extended-A
			return true
		}
		if r >= 0xFB50 && r <= 0xFDFF { // Arabic Presentation Forms-A
			return true
		}
		if r >= 0xFE70 && r <= 0xFEFF { // Arabic Presentation Forms-B
			return true
		}
	}
	return false
}

// splitLinePadding 拆出前导/尾随空白，译文保留中文缩进与行尾空格
func splitLinePadding(s string) (indent, content, trail string) {
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r != ' ' && r != '\t' {
			break
		}
		i += size
	}
	indent = s[:i]
	rest := s[i:]
	end := len(rest)
	for end > 0 {
		r, size := utf8.DecodeLastRuneInString(rest[:end])
		if r != ' ' && r != '\t' {
			break
		}
		end -= size
	}
	return indent, rest[:end], rest[end:]
}

func (d *Dict) translateLine(chinese, field string) string {
	indent, content, trail := splitLinePadding(chinese)
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return chinese // 空行/纯空白行原样保留
	}

	var body string
	if m := wholeParenRe.FindStringSubmatch(trimmed); m != nil {
		inner := strings.TrimSpace(m[1])
		if t := d.lookup(inner, field); t != "" {
			body = wrapInParen(t)
		} else {
			d.recordMiss(inner)
			body = "[" + inner + "]"
		}
		return indent + body + trail
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
		return indent + b.String() + trail
	}

	return indent + d.translatePlain(trimmed, field) + trail
}

func (d *Dict) translatePlain(chinese, field string) string {
	trimmed := strings.TrimSpace(chinese)
	if trimmed == "" {
		return ""
	}

	label, value, _, ok := splitByColon(trimmed)
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
		if translatedValue == "" {
			return displayLabel
		}
		// 有值时冒号后统一加空格：Fabric: 84.8%Acrylic
		return displayLabel + " " + translatedValue
	}

	if exact := d.lookup(trimmed, field); exact != "" {
		// 「成分」单独成行 → Composition:
		if (trimmed == "成分" || trimmed == "成份") && !strings.HasSuffix(exact, ":") {
			return exact + ":"
		}
		return exact
	}
	if comp := d.translateCompositionValue(trimmed, field); comp != "" {
		return comp
	}
	return d.translatePlainText(trimmed, field)
}

func splitByColon(s string) (label, value string, spaceAfter, ok bool) {
	idx := -1
	for i, r := range s {
		if r == '：' || r == ':' {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return "", "", false, false
	}
	label = strings.TrimSpace(s[:idx])
	_, size := utf8.DecodeRuneInString(s[idx:])
	rest := s[idx+size:]
	spaceAfter = len(rest) > 0 && (rest[0] == ' ' || rest[0] == '\t')
	value = strings.TrimSpace(rest)
	return label, value, spaceAfter, true
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
		if len(sub) < 4 {
			return match
		}
		pct, space, name := sub[1], sub[2], strings.TrimSpace(sub[3])
		if field == "arabic" && space == "" {
			space = " " // 阿语 % 后与材质名分开，利于渲染
		}
		token := pct + "%" // 所有语言：百分号在数字右边
		if t := d.lookup(name, field); t != "" && !strings.HasPrefix(t, "[") {
			return token + space + t
		}
		d.recordMiss(name)
		return token + space + name
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
