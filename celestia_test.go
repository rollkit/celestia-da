package celestia_test

import (
	"testing"

	"github.com/rollkit/celestia-da"
	"github.com/rollkit/go-da/test"
)

func TestCelestiaDA(t *testing.T) {
	da := &celestia.CelestiaDA{}
	test.RunDATestSuite(t, da)
}
