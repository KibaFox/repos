package repos_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"

	. "gitlab.com/KibaFox/repos/internal/repos"
)

var _ = Describe("Import", func() {
	It("walks directories and get collect repository paths and urls", func() {
		dir := importSetupRepos()
		defer cleanRepos(dir)

		repos, err := FromPath(dir)
		Expect(err).ShouldNot(HaveOccurred())

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
				URL:  "git@example.tld/KibaFox/test",
			},
			{
				Path: path.Join(h, "git.fqdn", "kiba", "spike"),
				URL:  "git@some.fqdn/KibaFox/spike",
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
~/git.fqdn/kiba/test     git@example.tld/KibaFox/test
~/git.fqdn/kiba/spike    git@some.fqdn/KibaFox/spike
~/git.fqdn/kira/dotfiles git@github.com/KiraFox/dotfiles
~/git.fqdn/kira/klok     git@github.com/KiraFox/klok
`))

		repos := parseSimple(bytes.NewReader(buf.Bytes()))
		Expect(repos).Should(ConsistOf(data))
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

	for _, d := range data {
		Expect(os.MkdirAll(d.path, 0755)).To(Succeed())

		repo, err := git.PlainInit(d.path, false)
		Expect(err).ToNot(HaveOccurred())

		if len(d.url) > 0 {
			_, err = repo.CreateRemote(&config.RemoteConfig{
				Name: "origin",
				URLs: []string{d.url},
			})
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
