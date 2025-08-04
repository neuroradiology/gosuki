// mod-gh-stars
// This gosuki module allows access to the stars and lists from the user's Github
// profile.

// # Github Personal Token:
// This module requires a github personal token to be able to access the user's profile.
// See full documentation at: https://gosuki.net/docs/configuration/config-file/#github-stars
package stars

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/blob42/gosuki"
	"github.com/blob42/gosuki/pkg/config"
	"github.com/blob42/gosuki/pkg/events"
	"github.com/blob42/gosuki/pkg/logging"
	"github.com/blob42/gosuki/pkg/modules"
	"github.com/blob42/gosuki/pkg/watch"

	"github.com/google/go-github/v66/github"
	"golang.org/x/time/rate"
)

var (
	GHToken  string
	Config   *StarFetcherConfig
	log      = logging.GetLogger(ModID)
	repoChan chan *github.StarredRepository
	errChan  = make(chan error, 10)
	wg       sync.WaitGroup
	sfModel  *StarFetcherModel
)

const (
	ModID = "github-stars"

	// default sync interval in seconds
	DefaultSyncInterval = 6 * time.Hour

	pageSize = 200
)

type StarFetcherConfig struct {
	GithubToken  string        `toml:"github-token" mapstructure:"github-token"`
	SyncInterval time.Duration `toml:"sync-interval" mapstructure:"sync-interval"`
}

func NewStarFetcherConfig() *StarFetcherConfig {
	return &StarFetcherConfig{
		SyncInterval: DefaultSyncInterval,
	}
}

type StarFetcherModel struct {
	// Github API token

	ctx context.Context

	token string
	gh    *github.Client
}

// This is the module struct. Used to implement module interface
type StarFetcher struct{}

// NOTE: use Init() to obtain context config params and store them in a model object
func (sf *StarFetcher) Init(ctx *modules.Context) error {
	sfModel.ctx = ctx.Context

	// Check if the environment variable GH_API_TOKEN is set
	if len(Config.GithubToken) == 0 {
		if sfModel.token = os.Getenv("GS_GITHUB_TOKEN"); sfModel.token == "" {
			return &modules.ErrModDisabled{Err: modules.ErrMissingCredentials}
		}
	} else {
		sfModel.token = Config.GithubToken
	}

	sfModel.gh = github.NewClient(nil).WithAuthToken(sfModel.token)

	return nil
}

func (sf StarFetcher) ModInfo() modules.ModInfo {

	return modules.ModInfo{
		ID: modules.ModID(ModID),
		New: func() modules.Module {
			return &StarFetcher{}
		},
	}
}

func (sf *StarFetcher) GetStarredRepos(done *sync.WaitGroup) {
	defer done.Done()
	ctx := sfModel.ctx
	gh := sfModel.gh

	opts := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{PerPage: pageSize},
	}

	limiter := rate.NewLimiter(rate.Every(time.Second), 10)

	total := 0
	curPage := 0
	for {
		log.Debug("fetching starred repos", "page", opts.Page)

		err := limiter.Wait(ctx)
		if err != nil {
			errChan <- err
		}

		repos, resp, err := gh.Activity.ListStarred(ctx, "", opts)
		if err != nil {
			errChan <- err
			return
		}

		log.Info("github ratelimit", "limit", resp.Rate.Limit, "remainging", resp.Rate.Remaining)

		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, repo := range repos {
				repoChan <- repo
			}
		}()

		total += len(repos)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage

		curPage++
	}

	wg.Wait()
	close(repoChan)
	errChan <- nil
}

func (sf *StarFetcher) Fetch() ([]*gosuki.Bookmark, error) {
	gh := sfModel.gh
	ctx := sfModel.ctx

	waitStars := &sync.WaitGroup{}
	log.Info("fetching stars")
	bookmarks := make([]*gosuki.Bookmark, 0, 100)

	// figuring out the total number of stars the user has
	// https://stackoverflow.com/questions/30636798/get-user-total-starred-count-using-github-api-v3

	opts := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{PerPage: 1, Page: 0},
	}
	_, resp, err := gh.Activity.ListStarred(ctx, "", opts)
	if err != nil {
		return nil, err
	}

	nStars := resp.LastPage

	// use a channel big enough to hold all stars and avoid blocking repoChan
	repoChan = make(chan *github.StarredRepository, nStars+pageSize)

	waitStars.Add(1)
	go sf.GetStarredRepos(waitStars)

	count := 0

	for repo := range repoChan {
		//DEBUG:
		// pretty.Println(repo)

		bk := gosuki.Bookmark{Module: string(sf.ModInfo().ID)}

		if repo.Repository.CloneURL != nil {
			bk.URL = *repo.Repository.CloneURL
		}

		if repo.Repository.Description != nil {
			bk.Desc = *repo.Repository.Description
		}

		if repo.Repository.Name != nil {
			bk.Title = *repo.Repository.Name
		}

		bk.Tags = repo.Repository.Topics

		bookmarks = append(bookmarks, &bk)
		count++

		go func() {
			events.TUIBus <- events.ProgressUpdateMsg{
				ID:           ModID,
				Instance:     nil,
				CurrentCount: uint(count),
				Total:        uint(len(bookmarks)),
			}
		}()
	}

	if err = <-errChan; err != nil {
		return nil, fmt.Errorf("fetching stars: %w", err)
	}

	waitStars.Wait()

	// bookmarks will be handled by the module runner
	return bookmarks, nil
}

// Interval at which the module should be run
func (sf StarFetcher) Interval() time.Duration {
	return Config.SyncInterval
}

func init() {
	sfModel = &StarFetcherModel{}

	// Custom Config
	Config = NewStarFetcherConfig()
	config.RegisterConfigurator(ModID, config.AsConfigurator(Config))
	modules.RegisterModule(&StarFetcher{})
}

// interface guards
var _ watch.Poller = (*StarFetcher)(nil)
var _ modules.Initializer = (*StarFetcher)(nil)
