package ff_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ff Suite")
}
