package scraper

import (
	"context"
	"sync"
	"time"

	"github.com/keircn/karu/internal/config"
)

type EpisodeJob struct {
	ShowID   string
	Episode  string
	Priority int
}

type LoadResult struct {
	ShowID    string
	Episode   string
	Qualities *QualityChoice
	Error     error
}

type ConcurrentLoader struct {
	workers    int
	jobs       chan EpisodeJob
	results    chan LoadResult
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	videoCache *Cache
}

func NewConcurrentLoader(workers int) *ConcurrentLoader {
	ctx, cancel := context.WithCancel(context.Background())

	loader := &ConcurrentLoader{
		workers:    workers,
		jobs:       make(chan EpisodeJob, workers*2),
		results:    make(chan LoadResult, workers*2),
		ctx:        ctx,
		cancel:     cancel,
		videoCache: NewCache(5 * time.Minute),
	}

	loader.start()
	return loader
}

func (cl *ConcurrentLoader) start() {
	for i := 0; i < cl.workers; i++ {
		cl.wg.Add(1)
		go cl.worker()
	}
}

func (cl *ConcurrentLoader) worker() {
	defer cl.wg.Done()

	for {
		select {
		case <-cl.ctx.Done():
			return
		case job := <-cl.jobs:
			result := cl.processJob(job)
			select {
			case cl.results <- result:
			case <-cl.ctx.Done():
				return
			}
		}
	}
}

func (cl *ConcurrentLoader) processJob(job EpisodeJob) LoadResult {
	cacheKey := generateCacheKey("quality", map[string]interface{}{
		"showId":  job.ShowID,
		"episode": job.Episode,
	})

	if cached, found := cl.videoCache.Get(cacheKey); found {
		return LoadResult{
			ShowID:    job.ShowID,
			Episode:   job.Episode,
			Qualities: cached.(*QualityChoice),
		}
	}

	qualities, err := GetAvailableQualities(job.ShowID, job.Episode)
	if err == nil && qualities != nil {
		cl.videoCache.Set(cacheKey, qualities)
	}

	return LoadResult{
		ShowID:    job.ShowID,
		Episode:   job.Episode,
		Qualities: qualities,
		Error:     err,
	}
}
func (cl *ConcurrentLoader) LoadEpisode(showID, episode string, priority int) {
	job := EpisodeJob{
		ShowID:   showID,
		Episode:  episode,
		Priority: priority,
	}

	select {
	case cl.jobs <- job:
	case <-cl.ctx.Done():
	}
}

func (cl *ConcurrentLoader) GetResult() *LoadResult {
	select {
	case result := <-cl.results:
		return &result
	case <-cl.ctx.Done():
		return nil
	}
}

func (cl *ConcurrentLoader) GetResultTimeout(timeout time.Duration) *LoadResult {
	select {
	case result := <-cl.results:
		return &result
	case <-time.After(timeout):
		return nil
	case <-cl.ctx.Done():
		return nil
	}
}

func (cl *ConcurrentLoader) PreloadEpisodes(showID string, episodes []string, currentIndex int) {
	cfg, _ := config.Load()
	maxPreload := cfg.PreloadEpisodes
	if maxPreload <= 0 {
		maxPreload = 5
	}

	start := currentIndex
	end := currentIndex + maxPreload

	if end > len(episodes) {
		end = len(episodes)
	}

	for i := start; i < end; i++ {
		priority := maxPreload - (i - currentIndex)
		cl.LoadEpisode(showID, episodes[i], priority)
	}

	if currentIndex > 0 {
		prevStart := currentIndex - maxPreload/2
		if prevStart < 0 {
			prevStart = 0
		}

		for i := prevStart; i < currentIndex; i++ {
			priority := 1
			cl.LoadEpisode(showID, episodes[i], priority)
		}
	}
}
func (cl *ConcurrentLoader) Shutdown() {
	cl.cancel()
	cl.wg.Wait()
}

var globalLoader *ConcurrentLoader
var loaderOnce sync.Once

func GetGlobalLoader() *ConcurrentLoader {
	loaderOnce.Do(func() {
		cfg, _ := config.Load()
		workers := cfg.ConcurrentWorkers
		if workers <= 0 {
			workers = 4
		}
		globalLoader = NewConcurrentLoader(workers)
	})
	return globalLoader
}

func PreloadAdjacentEpisodes(showID string, episodes []string, currentEpisode string) {
	loader := GetGlobalLoader()

	currentIndex := -1
	for i, ep := range episodes {
		if ep == currentEpisode {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return
	}

	loader.PreloadEpisodes(showID, episodes, currentIndex)
}

func GetVideoURLConcurrent(showID, episode string, timeout time.Duration) (string, error) {
	loader := GetGlobalLoader()

	cacheKey := generateCacheKey("quality", map[string]interface{}{
		"showId":  showID,
		"episode": episode,
	})

	if cached, found := loader.videoCache.Get(cacheKey); found {
		qualities := cached.(*QualityChoice)
		if len(qualities.Options) > 0 {
			return qualities.Options[qualities.Default].URL, nil
		}
	}

	loader.LoadEpisode(showID, episode, 10)

	result := loader.GetResultTimeout(timeout)
	if result == nil {
		return GetVideoURL(showID, episode)
	}

	if result.Error != nil {
		return "", result.Error
	}

	if result.Qualities != nil && len(result.Qualities.Options) > 0 {
		return result.Qualities.Options[result.Qualities.Default].URL, nil
	}

	return GetVideoURL(showID, episode)
}
