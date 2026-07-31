package service

import "testing"

func TestDetectColPairs(t *testing.T) {
	headers := []string{"材质", "款号", "洗唛成分", "洗唛成分英文", "洗唛成分俄文", "洗唛成分阿语", "洗涤说明", "洗涤说明英文", "洗涤说明俄文", "洗涤说明阿语", "产地", "产地英文", "产地俄文", "产地阿语"}
	pairs := detectColPairs(headers)
	if len(pairs) != 9 {
		t.Fatalf("want 9 pairs, got %d: %+v", len(pairs), pairs)
	}
}
