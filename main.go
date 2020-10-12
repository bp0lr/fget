package main

// imports
import (
	"errors"
	"fmt"
	"os"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"crypto/tls"
	"bufio"
	"time"
	"math/rand"
	"path"
	"strings"
	

	flag "github.com/spf13/pflag"
)

const (
	QSIZE = 20
)

type Ln struct {
	filename 	string
	url      	string
	host		string
}

var (
	workersArg	      int
	HeaderArg         []string
	urlArg            string
	statusListArg     string
	proxyArg          string
	fingerPrintArg    string
	outputFileArg     string
	queryArg          string
	verboseArg        bool
	followRedirectArg bool
	useRandomAgentArg bool
	testHTTPArg       bool	
)

func newClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:    30,
		IdleConnTimeout: time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout: time.Second * 5,
		}).DialContext,
	}

	if proxyArg != "" {
		if p, err := url.Parse(proxyArg); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	if !followRedirectArg {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}


func main() {

	flag.StringArrayVarP(&HeaderArg, "header", "H", nil, "Add custom Headers to the request")
	flag.StringVarP(&urlArg, "url", "u", "", "The url to check")
	flag.StringVarP(&outputFileArg, "output", "o", "", "Save to folder. Default: create results folder which will include a folder for each target")
	flag.IntVarP(&workersArg, "workers", "w", 20, "Number of workers")
	flag.BoolVarP(&verboseArg, "verbose", "v", false, "Display extra info about what is going on")
	flag.BoolVarP(&followRedirectArg, "follow-redirect", "f", false, "Follow redirects (Default: false)")
	flag.StringVarP(&proxyArg, "proxy", "p", "", "Add a HTTP proxy")	
	flag.BoolVarP(&useRandomAgentArg, "random-agent", "r", false, "Set a random User Agent")

	flag.Parse()

	if workersArg > 100 {
		workersArg = 100
	}

	client := newClient()

	var links []string

	if len(urlArg) > 0 {
		links = append(links, urlArg)
	} else {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			links = append(links, s.Text())
		}
	}

	var wg sync.WaitGroup
	
	queue := make(chan Ln, QSIZE)

	for i := 0; i < workersArg; i++ {
		wg.Add(1)
		go worker(i+1, queue, &wg, client)
	}

	for _, link := range links {
		
		u, err := url.ParseRequestURI(link)
		if err != nil {
			if(verboseArg){
				fmt.Printf("[-] Invalid url: %s\n", link)
			}
		}

		_, fileName := path.Split(link)

		link0:=Ln{filename: fileName, url: link, host: u.Host}
		
		queue <- link0
	}

	close(queue)
	wg.Wait()
}

func worker(index int, queue <-chan Ln, wg *sync.WaitGroup, client *http.Client) {
	
	defer wg.Done()
	for link := range queue {
		if(verboseArg){
			fmt.Printf("[+] Worker %d, downloading %s\n", index, link.url)
		}
		bytes, err := fetch(link.url, client)

		if err != nil {
			if(verboseArg){
				fmt.Println(err)
			}
			continue
		}

		var err_write error
		var fullPath string

		//save to file
		if(outputFileArg != ""){
			fullPath=path.Join(outputFileArg, link.filename)
			err_write = ioutil.WriteFile(fullPath, bytes, 0644)
		}else{

			//create if don't exists.
			if _, err_results := os.Stat("results"); os.IsNotExist(err_results) {
				os.Mkdir("results", 0644)
			}else{

			}

			fullPath=path.Join("results", link.host)

			//create folder for target
			if _, err_folder := os.Stat(fullPath); os.IsNotExist(err_folder) {
				os.Mkdir(fullPath, 0644)
			}

			fullPath=path.Join(fullPath, link.filename)
			err_write = ioutil.WriteFile(fullPath, bytes, 0644)
		}

		if err_write != nil {
			if(verboseArg){
				fmt.Printf("[-] error: %v\n", err_write)
			}
		} else {
			if(verboseArg){
				fmt.Printf("[+] file: %v, size: %v\n", link.filename, len(bytes))
			}
		}
	
	}
}

// Fetch function
func fetch(url string, client *http.Client) ([]byte, error) {
	
	//resp, err := http.Get(url)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		if verboseArg {
			fmt.Printf("[-] Error: %v\n", err)
		}
		return nil, errors.New("[-] Failed to fetch " + url)
	}

	if useRandomAgentArg {
		req.Header.Set("User-Agent", getUserAgent())
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; fget/1.0)")
	}

	// add headers to the request
	for _, h := range HeaderArg {
		parts := strings.SplitN(h, ":", 2)

		if len(parts) != 2 {
			continue
		}
		req.Header.Set(parts[0], parts[1])
	}

	// send the request
	resp, err := client.Do(req)
	if err != nil {
		if verboseArg {
			fmt.Printf("[-] Error: %v\n", err)
		}

		return nil, errors.New("[-] Failed to parse " + url)
	}
	defer resp.Body.Close()

	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if(verboseArg){
			return nil, errors.New("[-] Failed to parse " + url)
		}
	}
	return body, nil
}

func getUserAgent() string {
	payload := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 8_4_1 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12H321 Safari/600.1.4",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (iPad; CPU OS 7_1_2 like Mac OS X) AppleWebKit/537.51.2 (KHTML, like Gecko) Version/7.0 Mobile/11D257 Safari/9537.53",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(payload))

	pick := payload[randomIndex]

	return pick
}