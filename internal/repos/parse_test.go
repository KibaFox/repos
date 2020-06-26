package repos_test

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/kibafox/repos/internal/errs"
	. "gitlab.com/kibafox/repos/internal/repos"
)

var _ = Describe("Parse", func() {
	It("Should ignore comments", func() {
		config := "# this is a comment"
		repos := parseSimple(strings.NewReader(config))

		Expect(repos).Should(HaveLen(0))
	})

	It("Should ignore empty lines", func() {
		config := `# config w/ empty lines

`
		repos := parseSimple(strings.NewReader(config))

		Expect(repos).Should(HaveLen(0))
	})

	It("Should ignore lines with just tabs or spaces", func() {
		config := "# config w/ empty lines\n\t\t\n  \n"

		repos := parseSimple(strings.NewReader(config))

		Expect(repos).Should(HaveLen(0))
	})

	It("Should parse repos where PATH and URL are separated by space", func() {
		config := "/home/user/proj/test git@gitlab.com/user/test"

		repos := parseSimple(strings.NewReader(config))

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

		repos := parseSimple(strings.NewReader(config))

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

		repos := parseSimple(strings.NewReader(config))

		Expect(repos).Should(ConsistOf(
			Repo{
				Path: home + "/proj/test",
				URL:  "git@gitlab.com/user/test",
			},
		))
	})

	It("Allows any amount spaces and tabs between PATH and URL", func() {
		config := "/home/user/proj/test \t \t git@gitlab.com/user/test"

		repos := parseSimple(strings.NewReader(config))

		Expect(repos).Should(ConsistOf(
			Repo{
				Path: "/home/user/proj/test",
				URL:  "git@gitlab.com/user/test",
			},
		))
	})

	It("Skips when missing a URL", func() {
		config := "/home/user/proj/test"

		repos := parseErr(strings.NewReader(config), errs.ErrParseLine)

		Expect(repos).Should(HaveLen(0))
	})

	It("Skips when more fields than PATH and URL are given", func() {
		config := "/home/user/proj/test git@gitlab.com/user/test git@github.com"

		repos := parseErr(strings.NewReader(config), errs.ErrParseLine)

		Expect(repos).Should(HaveLen(0))
	})
})

// parseSimple will do a simple parse, expecting it to complete successfully.
func parseSimple(reader io.Reader) []Repo {
	var (
		repos []Repo
		err   error
		errs  = make(chan error, 1)
	)

	go func() {
		for err := range errs {
			log.Println(fmt.Errorf("sync: %w", err))
		}
	}()

	go func() {
		repos, err = Parse(reader, errs)
	}()

	Consistently(errs).ShouldNot(Receive())
	Expect(err).ShouldNot(HaveOccurred())

	return repos
}

// parseErr will expect one type of error to be pushed to the error channel.
func parseErr(reader io.Reader, err error) []Repo {
	var (
		repos    []Repo
		parseErr error
		errChan  = make(chan error, 1)
	)

	go func() {
		repos, parseErr = Parse(reader, errChan)
	}()

	for e := range errChan {
		log.Println(fmt.Errorf("parse: %w", e))
		Expect(errors.Unwrap(e)).Should(Equal(err))
	}

	Expect(parseErr).Should(HaveOccurred())

	return repos
}
