package repos_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/kibafox/repos/internal/git"
	. "gitlab.com/kibafox/repos/internal/repos"
)

var _ = Describe("Import", func() {
	It("walks directories and get collect repository paths and urls", func() {
		dir := importSetupRepos()
		defer cleanRepos(dir)

		repos := fromPathSimple(dir)

		Expect(repos).Should(ConsistOf(
			Repo{
				Path: path.Join(dir, "git.fqdn", "kiba", "dotfiles"),
				URL:  "git@gitlab.com/KibaFox/dotfiles",
			},
			Repo{
				Path: path.Join(dir, "git.fqdn", "kira", "dotfiles"),
				URL:  "git@github.com/KiraFox/dotfiles",
			},
			Repo{
				Path: path.Join(dir, "git.fqdn", "kira", "klok"),
				URL:  "git@github.com/KiraFox/klok",
			},
			Repo{
				Path: path.Join(dir, "git.fqdn", "kiba", "test"),
				URL:  "",
			},
			Repo{
				Path: path.Join(dir, "git.fqdn", "kiba", "spike"),
				URL:  "",
			},
		))
	})

	It("writes repositories that can be parsed", func() {
		h := home()

		data := []Repo{
			{
				Path: path.Join(h, "git.fqdn", "kiba", "dotfiles"),
				URL:  "git@gitlab.com/KibaFox/dotfiles",
			},
			{
				Path: path.Join(h, "git.fqdn", "kiba", "test"),
				URL:  "",
			},
			{
				Path: path.Join(h, "git.fqdn", "kiba", "spike"),
				URL:  "",
			},
			{
				Path: path.Join(h, "git.fqdn", "kira", "dotfiles"),
				URL:  "git@github.com/KiraFox/dotfiles",
			},
			{
				Path: path.Join(h, "git.fqdn", "kira", "klok"),
				URL:  "git@github.com/KiraFox/klok",
			},
		}

		var buf bytes.Buffer
		writer := bufio.NewWriter(&buf)
		Expect(WriteRepos(data, writer)).To(Succeed())
		writer.Flush()

		Expect(buf.String()).Should(
			Equal(`~/git.fqdn/kiba/dotfiles git@gitlab.com/KibaFox/dotfiles
# Could not find remote origin for local repository: ~/git.fqdn/kiba/test
# Could not find remote origin for local repository: ~/git.fqdn/kiba/spike
~/git.fqdn/kira/dotfiles git@github.com/KiraFox/dotfiles
~/git.fqdn/kira/klok     git@github.com/KiraFox/klok
`))

		repos := parseSimple(bytes.NewReader(buf.Bytes()))
		Expect(repos).Should(ConsistOf(
			Repo{
				Path: path.Join(h, "git.fqdn", "kiba", "dotfiles"),
				URL:  "git@gitlab.com/KibaFox/dotfiles",
			},
			Repo{
				Path: path.Join(h, "git.fqdn", "kira", "dotfiles"),
				URL:  "git@github.com/KiraFox/dotfiles",
			},
			Repo{
				Path: path.Join(h, "git.fqdn", "kira", "klok"),
				URL:  "git@github.com/KiraFox/klok",
			},
		))
	})
})

type testrepo struct {
	path string
	url  string
}

func importSetupRepos() (dir string) {
	Expect(os.MkdirAll("testdata", 0755)).To(Succeed())

	dir, err := ioutil.TempDir("testdata", "test_import_repos")
	Expect(err).ToNot(HaveOccurred())

	data := []testrepo{
		{
			path: path.Join(dir, "git.fqdn", "kiba", "dotfiles"),
			url:  "git@gitlab.com/KibaFox/dotfiles",
		},
		{
			path: path.Join(dir, "git.fqdn", "kiba", "test"),
			url:  "",
		},
		{
			path: path.Join(dir, "git.fqdn", "kiba", "spike"),
			url:  "",
		},
		{
			path: path.Join(dir, "git.fqdn", "kira", "dotfiles"),
			url:  "git@github.com/KiraFox/dotfiles",
		},
		{
			path: path.Join(dir, "git.fqdn", "kira", "klok"),
			url:  "git@github.com/KiraFox/klok",
		},
	}

	ctx := context.Background()

	for _, d := range data {
		Expect(os.MkdirAll(d.path, 0755)).To(Succeed())

		Expect(git.Run(ctx, "-C", d.path, "init")).To(Succeed())

		if len(d.url) > 0 {
			Expect(git.Run(ctx, "-C", d.path,
				"remote", "add", "origin", d.url)).To(Succeed())
		}

		Expect(err).ToNot(HaveOccurred())
	}

	return dir
}

func cleanRepos(dir string) {
	Expect(os.RemoveAll(dir)).To(Succeed())
}

func home() string {
	home, err := os.UserHomeDir()
	Expect(err).ToNot(HaveOccurred())

	return home
}

// fromPathSimple will do a simple import; expects it to complete successfully.
func fromPathSimple(path string) []Repo {
	var (
		repos       []Repo
		err         error
		errs        = make(chan error, 1)
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	)

	defer cancel()

	go func() {
		for err := range errs {
			log.Println(fmt.Errorf("import: %w", err))
		}
	}()

	go func() {
		repos, err = FromPath(ctx, path, errs)
	}()

	Consistently(errs).ShouldNot(Receive())
	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.Err()).ToNot(HaveOccurred())

	return repos
}
