package repos_test

import (
	"os"
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

	It("Should ignore lines with just tabs or spaces", func() {
		config := "# config w/ empty lines\n\t\t\n  \n"

		repos, err := Parse(strings.NewReader(config))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(repos).Should(HaveLen(0))
	})

	It("Should parse repos where PATH and URL are separated by space", func() {
		config := "/home/user/proj/test git@gitlab.com/user/test"

		repos, err := Parse(strings.NewReader(config))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(repos).Should(ConsistOf(
			Repo{
				Path: "/home/user/proj/test",
				URL:  "git@gitlab.com/user/test",
			},
		))
	})

	It("Should parse multiple repos", func() {
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

	It("Expands paths starting with `~/` to the user's home directory", func() {
		config := "~/proj/test git@gitlab.com/user/test"

		home, err := os.UserHomeDir()
		Expect(err).ShouldNot(HaveOccurred())

		repos, err := Parse(strings.NewReader(config))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(repos).Should(ConsistOf(
			Repo{
				Path: home + "/proj/test",
				URL:  "git@gitlab.com/user/test",
			},
		))
	})

	It("Allows any amount spaces and tabs between PATH and URL", func() {
		config := "/home/user/proj/test \t \t git@gitlab.com/user/test"

		repos, err := Parse(strings.NewReader(config))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(repos).Should(ConsistOf(
			Repo{
				Path: "/home/user/proj/test",
				URL:  "git@gitlab.com/user/test",
			},
		))
	})

	It("Skips when missing a URL", func() {
		config := "/home/user/proj/test"

		repos, err := Parse(strings.NewReader(config))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(repos).Should(HaveLen(0))
	})

	It("Skips when more fields than PATH and URL are given", func() {
		config := "/home/user/proj/test git@gitlab.com/user/test git@github.com"

		repos, err := Parse(strings.NewReader(config))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(repos).Should(HaveLen(0))
	})
})
