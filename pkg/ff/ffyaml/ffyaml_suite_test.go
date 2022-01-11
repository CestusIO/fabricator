package ffyaml_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFfyaml(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ffyaml Suite")
}
