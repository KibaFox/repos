package repos_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/kibafox/repos/internal/errs"
	"gitlab.com/kibafox/repos/internal/git"
	. "gitlab.com/kibafox/repos/internal/repos"
)

var _ = Describe("Sync", func() {
	var (
		repos []Repo
		dir   string
	)

	BeforeEach(func() {
		repos, dir = syncSetupRepos()
		log.Println(dir)
	})

	AfterEach(func() {
		cleanRepos(dir)
	})

	It("clones remote repositories initially", func() {
		syncSimple(repos)

		for _, r := range repos {
			Expect(r.Path).Should(BeADirectory())
			Expect(path.Join(r.Path, "README.md")).Should(BeARegularFile())
		}
	})

	It("pulls remote repositories after initial clone", func() {
		syncSimple(repos)

		By("Committing a CONTRIBUTING.md file to the remote repos")
		expectedHeads := make(map[string]string)
		for _, r := range repos {
			makeCommit(
				r.URL,
				"CONTRIBUTING.md",
				"# Contributing\n\nTODO\n",
				"Add CONTRIBUTING.md")

			expectedHeads[r.Path] = repoHeadHash(r.URL)
		}

		syncSimple(repos)

		for _, r := range repos {
			Expect(path.Join(r.Path, "CONTRIBUTING.md")).
				Should(BeARegularFile())
		}

		Expect(expectedHeads).Should(Equal(localHeadHashes(repos)))
	})

	It("does nothing when there are no updates", func() {
		syncSimple(repos)
		before := localHeadHashes(repos)

		syncSimple(repos)
		after := localHeadHashes(repos)

		Expect(before).Should(Equal(after))
	})

	It("only fetches when there are remote commits without conflicts", func() {
		syncSimple(repos)

		By("Committing a CONTRIBUTING.md file to the remote repos")
		remoteHeads := make(map[string]string)
		for _, r := range repos {
			makeCommit(
				r.URL,
				"CONTRIBUTING.md",
				"# Contributing\n\nTODO\n",
				"Add CONTRIBUTING.md")
			remoteHeads[r.Path] = repoHeadHash(r.URL)
		}

		By("Committing a TODO file to the local repos")
		localHeads := make(map[string]string)
		for _, r := range repos {
			makeCommit(
				r.Path,
				"TODO",
				"- write code\n- ???\n- profit\n",
				"Add TODO")

			localHeads[r.Path] = repoHeadHash(r.Path)
		}

		syncErr(repos, errs.ErrGit)
		after := localHeadHashes(repos)

		Expect(remoteHeads).ShouldNot(Equal(localHeads))
		Expect(localHeads).Should(Equal(after))

		By("Ensuring we fetched the remote commit")
		ctx := context.Background()
		for path, head := range remoteHeads {
			Expect(git.Run(ctx, "-C", path, "rev-parse", head)).To(Succeed())
		}
	})

	It("fetches when local has unstaged changes", func() {
		syncSimple(repos)
		before := localHeadHashes(repos)

		By("Committing an update to the README.md file to the remote repos")
		remoteHeads := make(map[string]string)
		for _, r := range repos {
			makeCommit(
				r.URL,
				"README.md",
				"# Readme\n\ntest 123\n",
				"Update README.md")

			remoteHeads[r.Path] = repoHeadHash(r.URL)
		}

		By("Modifying the README.md the local working directory")
		content := "- write code\n- ???\n- profit\n"
		for _, r := range repos {
			path := path.Join(r.Path, "README.md")
			Expect(ioutil.WriteFile(path, []byte(content), 0600)).To(Succeed())
		}

		syncErr(repos, errs.ErrGit)

		By("Verifying we only fetched since there were changes in the worktree")
		after := localHeadHashes(repos)
		Expect(remoteHeads).ShouldNot(Equal(after))
		Expect(before).Should(Equal(after))

		By("Ensuring we fetched to remote commit")
		ctx := context.Background()
		for path, head := range remoteHeads {
			Expect(git.Run(ctx, "-C", path, "rev-parse", head)).To(Succeed())
		}
	})

	It("fetches when there are items staged for commit", func() {
		syncSimple(repos)
		before := localHeadHashes(repos)

		By("Committing an update to the README.md file to the remote repos")
		remoteHeads := make(map[string]string)
		for _, r := range repos {
			makeCommit(
				r.URL,
				"README.md",
				"# Readme\n\ntest 123\n",
				"Update README.md")
			remoteHeads[r.Path] = repoHeadHash(r.URL)
		}

		By("Staging a change to the README.md in the local working directory")
		content := "- write code\n- ???\n- profit\n"

		ctx := context.Background()

		for _, r := range repos {
			path := path.Join(r.Path, "README.md")
			Expect(ioutil.WriteFile(path, []byte(content), 0600)).To(Succeed())
			Expect(git.Run(ctx, "-C", r.Path, "add", "README.md")).To(Succeed())
		}

		syncErr(repos, errs.ErrGit)

		By("Verifying we still pulled since there were no conflicts")
		after := localHeadHashes(repos)
		Expect(remoteHeads).ShouldNot(Equal(after))
		Expect(before).Should(Equal(after))

		By("Ensuring we fetched to remote commit")
		for path, head := range remoteHeads {
			Expect(git.Run(ctx, "-C", path, "rev-parse", head)).To(Succeed())
		}
	})
})

// syncSimple will do a simple sync, expecting it to complete successfully.
func syncSimple(repos []Repo) {
	var (
		err         error
		errs        = make(chan error, 1)
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	)

	defer cancel()

	go func() {
		for err := range errs {
			log.Println(fmt.Errorf("sync: %w", err))
		}
	}()

	go func() {
		err = Sync(ctx, repos, errs)
	}()

	Consistently(errs).ShouldNot(Receive())
	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.Err()).ToNot(HaveOccurred())
}

// syncErr will expect one type of error to be pushed to the error channel.
func syncErr(repos []Repo, err error) {
	var (
		syncErr     error
		errChan     = make(chan error, 1)
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	)

	defer cancel()

	go func() {
		syncErr = Sync(ctx, repos, errChan)
	}()

	for e := range errChan {
		log.Println(fmt.Errorf("sync: %w", e))
		Expect(errors.Unwrap(e)).Should(Equal(err))
	}

	Expect(syncErr).Should(HaveOccurred())
	Expect(ctx.Err()).ToNot(HaveOccurred())
}

const author = "Jane McHacker <jane.mc.hacker@notld>"

// syncSetupRepos will setup a test directory with the following repos:
//
//     .../remote/kiba
//     .../remote/kira
//
// The remote repos will be initialized with a single commit containing a
// README.md. They will map to the following local repos:
//
//     .../kiba/local
//     .../kira/local
//
// The local repos will not be synced.
func syncSetupRepos() (repos []Repo, dir string) {
	Expect(os.MkdirAll("testdata", 0755)).To(Succeed())

	dir, err := ioutil.TempDir("testdata", "test_sync_repos")
	Expect(err).ToNot(HaveOccurred())

	repos = []Repo{
		{
			Path: path.Join(dir, "kiba", "local"),
			URL:  path.Join(dir, "remote", "kiba"),
		},
		{
			Path: path.Join(dir, "kira", "local"),
			URL:  path.Join(dir, "remote", "kira"),
		},
	}

	content := []byte("# README\n\nTODO\n")

	ctx := context.Background()

	for _, r := range repos {
		Expect(os.MkdirAll(r.URL, 0755)).To(Succeed())

		Expect(git.Run(ctx, "-C", r.URL, "init")).To(Succeed())

		readme := path.Join(r.URL, "README.md")
		Expect(ioutil.WriteFile(readme, content, 0600)).To(Succeed())

		Expect(git.Run(ctx, "-C", r.URL, "add", "README.md")).To(Succeed())

		Expect(git.Run(ctx, "-C", r.URL, "commit", "-m", "Add README.md",
			"--author", author)).To(Succeed())
	}

	return repos, dir
}

func repoHeadHash(path string) string {
	ctx := context.Background()

	hash, err := git.Out(ctx, "-C", path, "rev-parse", "HEAD")
	Expect(err).ToNot(HaveOccurred())

	return hash
}

func localHeadHashes(repos []Repo) (heads map[string]string) {
	heads = make(map[string]string)

	for _, r := range repos {
		heads[r.Path] = repoHeadHash(r.Path)
	}

	Expect(heads).Should(HaveLen(len(repos)))

	return heads
}

func makeCommit(repoPath, file, content, msg string) {
	ctx := context.Background()

	path := path.Join(repoPath, file)
	Expect(ioutil.WriteFile(path, []byte(content), 0600)).To(Succeed())

	Expect(git.Run(ctx, "-C", repoPath, "add", file)).To(Succeed())

	Expect(git.Run(ctx, "-C", repoPath,
		"commit", "-m", msg, "--author", author)).To(Succeed())
}
