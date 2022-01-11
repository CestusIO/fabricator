package ffpflag_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFfpflag(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ffpflag Suite")
}
