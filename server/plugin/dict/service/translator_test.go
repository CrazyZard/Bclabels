package service

import (
	"strings"
	"testing"
)

func TestWrapParenNoDouble(t *testing.T) {
	d := &Dict{Entries: map[string]map[string]string{
		"（薄膜除外）": {"english": "(Except film)", "russian": "кроме плёнки", "arabic": "(باستثناء فيلم)"},
		"成分":     {"english": "Composition"},
		"面料":     {"english": "Fabric"},
		"里料":     {"english": "Lining"},
		"聚酯纤维":   {"english": "Polyester"},
	}}

	got := d.TranslateText("（薄膜除外）", "english")
	if got != "(Except film)" {
		t.Fatalf("want (Except film), got %q", got)
	}
	got = d.TranslateText("(薄膜除外)", "english")
	if got != "(Except film)" {
		t.Fatalf("want (Except film), got %q", got)
	}

	src := "成分：\n面料：100%聚酯纤维\n              (薄膜除外)\n里料：100%聚酯纤维"
	got = d.TranslateText(src, "english")
	want := "Composition:\nFabric: 100%Polyester\n              (Except film)\nLining: 100%Polyester"
	if got != want {
		t.Fatalf("composition mismatch:\nwant:\n%s\ngot:\n%s", want, got)
	}
	if strings.Contains(got, "((") {
		t.Fatalf("still has double paren: %q", got)
	}
}

func TestUserCompositionSample(t *testing.T) {
	d := &Dict{Entries: map[string]map[string]string{
		"成分":         {"english": "Composition"},
		"面料":         {"english": "Fabric"},
		"里料":         {"english": "Lining"},
		"腈纶":         {"english": "Acrylic"},
		"绵羊毛":       {"english": "Sheep wool"},
		"罗纹和配料除外": {"english": "Except rib and accessories"},
		"聚酯纤维":     {"english": "Polyester"},
	}}

	src := "成分\n面料：84.8%腈纶\n           15.2%绵羊毛 \n （罗纹和配料除外）\n里料：100%聚酯纤维"
	got := d.TranslateText(src, "english")
	want := "Composition:\nFabric: 84.8%Acrylic\n           15.2%Sheep wool \n (Except rib and accessories)\nLining: 100%Polyester"
	if got != want {
		t.Fatalf("sample mismatch:\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func stripBidiMarks(s string) string {
	return strings.NewReplacer(
		"\u200E", "", "\u200F", "",
		"\u202A", "", "\u202B", "", "\u202C", "", "\u202D", "", "\u202E", "",
		"\u2066", "", "\u2067", "", "\u2068", "", "\u2069", "",
	).Replace(s)
}

func TestArabicPercentOrder(t *testing.T) {
	d := &Dict{Entries: map[string]map[string]string{
		"成分":     {"arabic": "المكونات"},
		"面料":     {"arabic": "النسيج"},
		"里料":     {"arabic": "البطانة"},
		"聚酯纤维": {"arabic": "البوليستر"},
		"其他纤维": {"arabic": "ألياف أخرى"},
		"腈纶":     {"arabic": "البولي أكريلونيتريل"},
		"绵羊毛":   {"arabic": "صوف الأغنام"},
		// 字典整句可能写成 % 在前
		"99.4%聚酯纤维": {"arabic": "%99.4البوليستر"},
	}}

	src := "成分：\n面料：99.4%聚酯纤维\n           0.6%其他纤维"
	got := d.TranslateText(src, "arabic")
	// 有阿语前缀的行用 %99.4；缩进续行保持 0.6%
	want := "المكونات:\nالنسيج: %99.4 البوليستر\n           0.6% ألياف أخرى"
	if got != want {
		t.Fatalf("arabic percent order:\nwant:\n%q\ngot:\n%q", want, got)
	}

	src2 := "成分\n面料：84.8%腈纶\n           15.2%绵羊毛 \n里料：100%聚酯纤维"
	got2 := d.TranslateText(src2, "arabic")
	want2 := "المكونات:\nالنسيج: %84.8 البولي أكريلونيتريل\n           15.2% صوف الأغنام \nالبطانة: %100 البوليستر"
	if got2 != want2 {
		t.Fatalf("arabic fabric sample:\nwant:\n%q\ngot:\n%q", want2, got2)
	}

	if got := normalizePercentOrder("%84.8 البولي و ٪15.2 صوف و 100 ٪"); got != "84.8% البولي و 15.2% صوف و 100%" {
		t.Fatalf("normalize got %q", got)
	}
}

func TestPreserveCompositionIndent(t *testing.T) {
	d := &Dict{Entries: map[string]map[string]string{
		"成分":   {"english": "Composition"},
		"面料":   {"english": "Fabric"},
		"里料":   {"english": "Lining"},
		"棉":     {"english": "Cotton"},
		"薄膜除外": {"english": "Except film"},
	}}
	src := "成分：\n面料：70%棉\n    (薄膜除外)\n\n里料：30%棉"
	got := d.TranslateText(src, "english")
	want := "Composition:\nFabric: 70%Cotton\n    (Except film)\n\nLining: 30%Cotton"
	if got != want {
		t.Fatalf("indent/format mismatch:\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestTranslateComposition(t *testing.T) {
	d := &Dict{Entries: map[string]map[string]string{
		"成分":     {"english": "Composition", "russian": "Состав", "arabic": "التركيب"},
		"面料":     {"english": "Shell", "russian": "Верх", "arabic": "القماش"},
		"里料":     {"english": "Lining", "russian": "Подкладка", "arabic": "البطانة"},
		"聚酯纤维":   {"english": "Polyester", "russian": "Полиэстер", "arabic": "بوليстер"},
		"中国制造":   {"english": "Made in China", "russian": "Сделано в Китае", "arabic": "صنع في الصين"},
		"不可干洗":   {"english": "Do not dry clean", "russian": "Не подвергать химчистке", "arabic": "لا تنظف جافاً"},
	}}

	src := "成分：\n面料：100%聚酯纤维\n              (薄膜除外)\n里料：100%聚酯纤维"
	en := d.TranslateText(src, "english")
	t.Log("EN:\n" + en)
	if en == src {
		t.Fatal("not translated")
	}

	origin := d.TranslateText("中国制造", "english")
	if origin != "Made in China" {
		t.Fatalf("got %q", origin)
	}
	wash := d.TranslateText("不可干洗", "english")
	if wash != "Do not dry clean" {
		t.Fatalf("got %q", wash)
	}
}
