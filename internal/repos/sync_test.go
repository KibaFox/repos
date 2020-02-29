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
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	. "gitlab.com/KibaFox/repos/internal/repos"
)

var _ = Describe("Sync", func() {
	var (
		repos  []Repo
		dir    string
		asJeff *git.CommitOptions
	)

	BeforeEach(func() {
		asJeff = &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Jeff McTester",
				Email: "jeff.mc.tester@notld",
			},
		}
		repos, dir = syncSetupRepos(asJeff)
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
		expectedHeads := make(map[string]plumbing.Hash)
		for _, r := range repos {
			expectedHeads[r.Path] = makeCommit(
				r.URL,
				"CONTRIBUTING.md",
				"# Contributing\n\nTODO\n",
				"Add CONTRIBUTING.md",
				asJeff)
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
		remoteHeads := make(map[string]plumbing.Hash)
		for _, r := range repos {
			remoteHeads[r.Path] = makeCommit(
				r.URL,
				"CONTRIBUTING.md",
				"# Contributing\n\nTODO\n",
				"Add CONTRIBUTING.md",
				asJeff)
		}

		By("Committing a TODO file to the local repos")
		localHeads := make(map[string]plumbing.Hash)
		for _, r := range repos {
			localHeads[r.Path] = makeCommit(
				r.Path,
				"TODO",
				"- write code\n- ???\n- profit\n",
				"Add TODO",
				asJeff)
		}

		syncErr(repos, ErrNonFastForwardUpdate)
		after := localHeadHashes(repos)

		Expect(remoteHeads).ShouldNot(Equal(localHeads))
		Expect(localHeads).Should(Equal(after))

		By("Ensuring we fetched the remote commit")
		for path, head := range remoteHeads {
			_, err := repoOpen(path).Object(plumbing.CommitObject, head)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("fetches when the local directory is dirty", func() {
		syncSimple(repos)
		before := localHeadHashes(repos)

		By("Committing an update to the README.md file to the remote repos")
		remoteHeads := make(map[string]plumbing.Hash)
		for _, r := range repos {
			remoteHeads[r.Path] = makeCommit(
				r.URL,
				"README.md",
				"# Readme\n\ntest 123\n",
				"Update README.md",
				asJeff)
		}

		By("Modifying the README.md the local working directory")
		content := "- write code\n- ???\n- profit\n"
		for _, r := range repos {
			path := path.Join(r.Path, "README.md")
			Expect(ioutil.WriteFile(path, []byte(content), 0600)).To(Succeed())
		}

		syncSimple(repos)

		By("Verifying we only fetched since there were changes in the worktree")
		after := localHeadHashes(repos)
		Expect(remoteHeads).ShouldNot(Equal(after))
		Expect(before).Should(Equal(after))

		By("Ensuring we fetched to remote commit")
		for path, head := range remoteHeads {
			_, err := repoOpen(path).Object(plumbing.CommitObject, head)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("pulls when there are staged items in the local directory", func() {
		syncSimple(repos)
		before := localHeadHashes(repos)

		By("Committing an update to the README.md file to the remote repos")
		remoteHeads := make(map[string]plumbing.Hash)
		for _, r := range repos {
			remoteHeads[r.Path] = makeCommit(
				r.URL,
				"README.md",
				"# Readme\n\ntest 123\n",
				"Update README.md",
				asJeff)
		}

		By("Staging a change to the README.md in the local working directory")
		content := "- write code\n- ???\n- profit\n"
		for _, r := range repos {
			path := path.Join(r.Path, "README.md")
			Expect(ioutil.WriteFile(path, []byte(content), 0600)).To(Succeed())
			_, err := repoWrk(repoOpen(r.Path)).Add("README.md")
			Expect(err).ToNot(HaveOccurred())
		}

		syncSimple(repos)

		By("Verifying we still pulled since there were no conflicts")
		after := localHeadHashes(repos)
		Expect(remoteHeads).ShouldNot(Equal(after))
		Expect(before).Should(Equal(after))

		By("Ensuring we fetched to remote commit")
		for path, head := range remoteHeads {
			_, err := repoOpen(path).Object(plumbing.CommitObject, head)
			Expect(err).ToNot(HaveOccurred())
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
func syncSetupRepos(opt *git.CommitOptions) (repos []Repo, dir string) {
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

	for _, r := range repos {
		Expect(os.MkdirAll(r.URL, 0755)).To(Succeed())

		repo, err := git.PlainInit(r.URL, false)
		Expect(err).ToNot(HaveOccurred())

		wrk := repoWrk(repo)

		readme := path.Join(r.URL, "README.md")
		Expect(ioutil.WriteFile(readme, content, 0600)).To(Succeed())

		_, err = wrk.Add("README.md")
		Expect(err).ToNot(HaveOccurred())

		_, err = wrk.Commit("Add README.md", opt)
		Expect(err).ToNot(HaveOccurred())
	}

	return repos, dir
}

func repoOpen(path string) (repo *git.Repository) {
	var err error
	repo, err = git.PlainOpen(path)
	Expect(err).ToNot(HaveOccurred())

	return repo
}

func repoWrk(repo *git.Repository) (wrk *git.Worktree) {
	var err error
	wrk, err = repo.Worktree()
	Expect(err).ToNot(HaveOccurred())

	return wrk
}

func repoHeadHash(repo *git.Repository) (hash plumbing.Hash) {
	head, err := repo.Head()
	Expect(err).ToNot(HaveOccurred())

	return head.Hash()
}

func localHeadHashes(repos []Repo) (heads map[string]plumbing.Hash) {
	heads = make(map[string]plumbing.Hash)

	for _, r := range repos {
		heads[r.Path] = repoHeadHash(repoOpen(r.Path))
	}

	Expect(heads).Should(HaveLen(len(repos)))

	return heads
}

func makeCommit(
	repoPath, file, content, msg string, opt *git.CommitOptions,
) (hash plumbing.Hash) {
	var err error

	repo := repoOpen(repoPath)
	wrk := repoWrk(repo)

	path := path.Join(repoPath, file)
	Expect(ioutil.WriteFile(path, []byte(content), 0600)).To(Succeed())

	_, err = wrk.Add(file)
	Expect(err).ToNot(HaveOccurred())

	hash, err = wrk.Commit(msg, opt)
	Expect(err).ToNot(HaveOccurred())

	return hash
}
