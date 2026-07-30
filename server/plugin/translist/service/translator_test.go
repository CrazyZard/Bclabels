package service

import (
	"strings"
	"testing"
)

func TestDetectColPairs(t *testing.T) {
	headers := []string{"材质", "款号", "洗唛成分", "洗唛成分英文", "洗唛成分俄文", "洗唛成分阿语", "洗涤说明", "洗涤说明英文", "洗涤说明俄文", "洗涤说明阿语", "产地", "产地英文", "产地俄文", "产地阿语"}
	pairs := detectColPairs(headers)
	if len(pairs) != 9 {
		t.Fatalf("want 9 pairs, got %d: %+v", len(pairs), pairs)
	}
}

func TestWrapParenNoDouble(t *testing.T) {
	d := &Dict{Entries: map[string]map[string]string{
		"（薄膜除外）": {"english": "(Except film)", "russian": "кроме плёнки", "arabic": "(باستثناء فيلم)"},
		"成分":       {"english": "Composition"},
		"面料":       {"english": "Fabric"},
		"里料":       {"english": "Lining"},
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
	want := "Composition:\nFabric: 100% Polyester\n(Except film)\nLining: 100% Polyester"
	if got != want {
		t.Fatalf("composition mismatch:\nwant:\n%s\ngot:\n%s", want, got)
	}
	if strings.Contains(got, "((") {
		t.Fatalf("still has double paren: %q", got)
	}
}

func TestTranslateComposition(t *testing.T) {
	d := &Dict{Entries: map[string]map[string]string{
		"成分":     {"english": "Composition", "russian": "Состав", "arabic": "التركيب"},
		"面料":     {"english": "Shell", "russian": "Верх", "arabic": "القماش"},
		"里料":     {"english": "Lining", "russian": "Подкладка", "arabic": "البطانة"},
		"聚酯纤维": {"english": "Polyester", "russian": "Полиэстер", "arabic": "بوليستر"},
		"中国制造": {"english": "Made in China", "russian": "Сделано в Китае", "arabic": "صنع في الصين"},
		"不可干洗": {"english": "Do not dry clean", "russian": "Не подвергать химчистке", "arabic": "لا تنظف جافاً"},
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
