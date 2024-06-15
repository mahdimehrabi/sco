package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	logger "m1-article-service/infrastructure/log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/nfnt/resize"
	"golang.org/x/time/rate"
)

const (
	numWorkers       = 10000
	rateLimit        = 1000 //maximum 1000 image download per second (iran network cant reach maximum but its possible in a server with good resources and internet)
	imageWidth       = 100
	downloadQueueCap = 100000
	proxyFetchURL    = "https://www.sslproxies.org/"
)

var petQueries = []string{
	"cute kittens", "puppies", "hamsters", "bunnies", "goldfish",
	"parrots", "turtles", "guinea pigs", "hedgehogs", "ferrets",
	"pet snakes", "pet lizards", "pet frogs", "pet spiders", "pet mice",
	"pet rats", "pet birds", "pet rabbits", "pet ducks", "pet chickens",
}

type Downloader interface {
	Download(pathsChan chan string)
}

type SearchEngine struct {
	Name       string
	SearchURL  string
	ResultAttr string
	Extractor  func(*colly.HTMLElement) string
}

var searchEngines = []SearchEngine{
	{"Google", "https://www.google.com/search?tbm=isch&q=%s", "img",
		func(element *colly.HTMLElement) string {
			return element.Attr("src")
		}},
	{"Bing", "https://www.bing.com/images/search?q=%s", "a.iusc", extractImageURLFromBing},
}

func extractImageURLFromBing(e *colly.HTMLElement) string {
	data := e.Attr("m")
	start := strings.Index(data, `"murl":"`)
	if start == -1 {
		return ""
	}
	start += len(`"murl":"`)
	end := strings.Index(data[start:], `"`)
	if end == -1 {
		return ""
	}
	return data[start : start+end]
}

type DownloadResizer struct {
	downloadQueue chan string
	saveDirectory string
	logger        logger.Logger
	count         uint64
	targetCount   uint64
	limiter       *rate.Limiter
	mtx           *sync.Mutex
	proxies       []string
	rand          *rand.Rand
	ctx           context.Context
	cancelCtx     context.CancelFunc
	proxy         bool
	resultChan    chan string
}

func NewDownloadResizer(saveDir string, targetCount uint64, lg logger.Logger, proxy bool) *DownloadResizer {
	s := rand.NewSource(time.Now().UnixNano())
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &DownloadResizer{
		downloadQueue: make(chan string, downloadQueueCap),
		saveDirectory: saveDir,
		targetCount:   targetCount,
		logger:        lg,
		limiter:       rate.NewLimiter(rate.Limit(rateLimit), rateLimit),
		mtx:           &sync.Mutex{},
		rand:          rand.New(rand.New(s)),
		ctx:           ctx,
		cancelCtx:     cancelFunc,
		proxy:         proxy,
	}
}

// Download: returns channel of file paths
func (d *DownloadResizer) Download(pathsChan chan string) {
	d.resultChan = pathsChan
	if d.proxy {
		//we refresh proxies for every download starting command because of awful connection quality of free proxies
		go func() {
			for {
				if d.count < d.targetCount {
					if err := d.fetchProxies(); err != nil {
						d.logger.Warning("failed to fetch new proxies proxies:" + err.Error())
						d.logger.Info("running program without proxies...")
						d.proxies = make([]string, 0)
					}
					time.Sleep(7 * time.Second)
				} else {
					return
				}
			}
		}()
	}
	// Start workers to process image URLs
	for i := 0; i < numWorkers; i++ {
		go d.worker()
	}

	c := colly.NewCollector(
		colly.Async(true),
	)
	c.AllowURLRevisit = true
	c.SetRequestTimeout(time.Second * 2)

loop:
	for {
		select {
		case <-d.ctx.Done():
			break loop
		default:
			engine := searchEngines[d.rand.Intn(len(searchEngines))]
			c.OnHTML(engine.ResultAttr, func(e *colly.HTMLElement) {
				imgURL := engine.Extractor(e)
				if imgURL != "" {
					d.downloadQueue <- imgURL
				}
			})
			query := petQueries[d.rand.Intn(len(petQueries))]
			searchURL := fmt.Sprintf(engine.SearchURL, url.QueryEscape(query))
			d.logger.Info(fmt.Sprintf("Scraping %s for '%s'...\n", engine.Name, strings.ReplaceAll(query, " ", "+")))

			if err := c.Visit(searchURL); err != nil {
				d.logger.Error(err)
				continue
			}
			c.Wait()
		}
	}
	close(d.downloadQueue)
	close(d.resultChan)

	d.logger.Info("Finished downloading and processing images.")

	return
}

func (d *DownloadResizer) worker() {

	for imgURL := range d.downloadQueue {
		// Wait for the rate limiter
		if err := d.limiter.Wait(context.Background()); err != nil {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		d.downloadAndResizeImage(imgURL)
	}
}

func (d *DownloadResizer) downloadAndResizeImage(imageUrl string) (err error) {
	filePath := fmt.Sprintf("%d.jpg", time.Now().UnixNano()+int64(d.rand.Intn(9999)))
	fullPath := filepath.Join(d.saveDirectory, filePath)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client := &http.Client{}
	if len(d.proxies) > 0 {
		proxyURL, err := url.Parse(d.getRandomProxy())
		if err != nil {
			return err
		}

		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", imageUrl, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return err
	}

	m := resize.Resize(imageWidth, 0, img, resize.Lanczos3)

	d.mtx.Lock()
	defer d.mtx.Unlock()
	if d.count >= d.targetCount {
		return
	}
	out, err := os.Create(fullPath)
	defer func() {
		out.Close()
		if err != nil {
			if err = os.Remove(fullPath); err != nil {
				d.logger.Warning(err.Error())
			}
		}
	}()
	if err != nil {
		return err
	}
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		return err
	}
	d.count++
	d.resultChan <- filePath
	d.logger.Info(fmt.Sprintf("downloaded %d images", d.count))
	if d.count == d.targetCount {
		d.cancelCtx()
	}

	return nil
}

func (d *DownloadResizer) getRandomProxy() string {
	return d.proxies[d.rand.Intn(len(d.proxies))]
}

func (d *DownloadResizer) fetchProxies() error {
	var proxies []string
	c := colly.NewCollector()

	c.OnHTML("table.table tr", func(e *colly.HTMLElement) {
		ip := e.ChildText("td:nth-child(1)")
		port := e.ChildText("td:nth-child(2)")
		if ip != "" && port != "" {
			proxy := fmt.Sprintf("http://%s:%s", ip, port)
			proxies = append(proxies, proxy)
		}
	})

	err := c.Visit(proxyFetchURL)
	if err != nil {
		return err
	}
	c.Wait()

	d.proxies = proxies
	return nil
}
