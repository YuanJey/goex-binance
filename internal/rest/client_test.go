package rest

import "testing"

func TestSign(t *testing.T) {
	const payload = "symbol=BTCUSDT&timestamp=1499827319559"
	const want = "ce4803920f488e828bc5b2adda557c0dc4ffed151e99432bd132f31d0299dad8"
	if got := sign("secret", payload); got != want {
		t.Fatalf("sign()=%s want %s", got, want)
	}
}
