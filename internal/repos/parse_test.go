package repos_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "gitlab.com/KibaFox/repos/internal/repos"
)

var _ = Describe("Parse", func() {
	It("Should ignore comments", func() {
		config := "# this is a comment"
		repos, err := Parse(strings.NewReader(config))

		Expect(err).ShouldNot(HaveOccurred())
		Expect(repos).Should(HaveLen(0))
	})

	It("Should ignore empty lines", func() {
		config := `# config w/ empty lines

`
		repos, err := Parse(strings.NewReader(config))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(repos).Should(HaveLen(0))
	})

	It("Should parse repos where PATH and URL are separated by space", func() {
		config := `# These repos are separated by a single space
/home/user/proj/test git@gitlab.com/user/test

/home/user/proj/asdf git@gitlab.com/user/asdf

/home/user/proj/qwer git@gitlab.com/user/qwer
`

		repos, err := Parse(strings.NewReader(config))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(repos).Should(ConsistOf(
			Repo{
				Path: "/home/user/proj/test",
				URL:  "git@gitlab.com/user/test",
			},
			Repo{
				Path: "/home/user/proj/asdf",
				URL:  "git@gitlab.com/user/asdf",
			},
			Repo{
				Path: "/home/user/proj/qwer",
				URL:  "git@gitlab.com/user/qwer",
			},
		))
	})
})
