package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	tok, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		panic("GITHUB_TOKEN required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tok},
	)
	tc := oauth2.NewClient(ctx, ts)
	cl := github.NewClient(tc)

	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       "closed",
	}

	// Stop at this PR number
	stopAt := 428
	includeCommits := false //true

	// Iterate over all PRs
Done:
	for {
		prs, resp, err := cl.PullRequests.List(ctx, "ohsu-comp-bio", "funnel", opt)
		if err != nil {
			panic(err)
		}

		// Iterate over PRs in this page.
		for _, pr := range prs {
			if *pr.Number == stopAt {
				break Done
			}
			if pr.MergedAt == nil {
				continue
			}

			fmt.Printf("- PR #%d %s\n", pr.GetNumber(), pr.GetTitle())

			if includeCommits {
				// Iterate over all commits in this PR.
				commitOpt := &github.ListOptions{PerPage: 100}
				for {

					commits, resp, err := cl.PullRequests.ListCommits(ctx, "ohsu-comp-bio", "funnel", pr.GetNumber(), commitOpt)
					if err != nil {
						panic(err)
					}

					// Iterate over commits in this page.
					for _, commit := range commits {
						sha := *commit.SHA
						msg := *commit.Commit.Message

						// Strip multiple lines (i.e. only take first line)
						if i := strings.Index(msg, "\n"); i != -1 {
							msg = msg[:i]
						}
						// Trim long lines
						if len(msg) > 90 {
							msg = msg[:90] + "..."
						}
						msg = strings.TrimSpace(msg)

						fmt.Printf("    - %s %s\n", sha, msg)
					}

					if resp.NextPage == 0 {
						break
					}
					commitOpt.Page = resp.NextPage
				}
				fmt.Println()
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}
