package g_test

import (
	"testing"

	"github.com/loov/allocview/internal/ui/g"
)

func TestHit(t *testing.T) {
	r := g.Rect{
		Min: g.V0,
		Max: g.V1,
	}

	if r.Test(r.TopLeft(), 0.1) != g.Inside|g.Top|g.Left {
		t.Error("Top Left")
	}
	if r.Test(r.TopRight(), 0.1) != g.Inside|g.Top|g.Right {
		t.Error("Top Right")
	}
	if r.Test(r.BottomLeft(), 0.1) != g.Inside|g.Bottom|g.Left {
		t.Error("Bottom Left")
	}
	if r.Test(r.BottomRight(), 0.1) != g.Inside|g.Bottom|g.Right {
		t.Error("Bottom Right")
	}

	if r.Test(r.TopCenter(), 0.1) != g.Inside|g.Top {
		t.Error("Top")
	}
	if r.Test(r.BottomCenter(), 0.1) != g.Inside|g.Bottom {
		t.Error("Bottom")
	}
	if r.Test(r.LeftCenter(), 0.1) != g.Inside|g.Left {
		t.Error("Left")
	}
	if r.Test(r.RightCenter(), 0.1) != g.Inside|g.Right {
		t.Error("Right")
	}
}
