package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"golang.org/x/net/html"
	"golang.org/x/time/rate"
)

var mirrorFlag bool

func init() {
	flag.BoolVar(&mirrorFlag, "mirror", false, "Mirror the remote directory structure")
}

// Function to format file size into human-readable format
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB (%d bytes)", float64(bytes)/float64(GB), bytes)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB (%d bytes)", float64(bytes)/float64(MB), bytes)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB (%d bytes)", float64(bytes)/float64(KB), bytes)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// convertLinks modifies links in an HTML file to point to local copies.
func convertLinks(html string, baseURL string) string {
	base, _ := url.Parse(baseURL)

	rgx := regexp.MustCompile(`(href|src)=["']([^"']+)["']`)
	html = rgx.ReplaceAllStringFunc(html, func(match string) string {
		parts := rgx.FindStringSubmatch(match)
		attr, link := parts[1], parts[2]

		u, err := url.Parse(link)
		if err != nil || u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
			return match
		}
		if u.Host != "" && u.Host != base.Host {
			return match
		}

		relPath := path.Join("mirrored_site", u.Path)
		return fmt.Sprintf(`%s="%s"`, attr, relPath)
	})

	return html
}
// shouldExclude checks if a URL should be excluded based on the -X flag
func shouldExclude(link string, excludes []string) bool {
	for _, exclude := range excludes {
		if strings.Contains(link, exclude) {
			return true
		}
	}
	return false
}
// mirrorWebsite recursively downloads and processes pages
// Higher-Level Function: Handles website mirroring
func mirrorWebsite(url string, rejectTypes string, acceptType string, recursive bool) error {
	startTime := time.Now()
	fmt.Printf("Starting website mirroring: %s\n", url)
	fmt.Printf("Started mirroring at: %s\n", startTime.Format("yyyy-mm-dd hr:min:sec"))

	// Create a directory to store the website
	domain := extractDomain(url)
	outputDir := fmt.Sprintf("%s_mirror", domain)
	os.MkdirAll(outputDir, os.ModePerm)

	// Fetch main page
	htmlContent, err := fetchHTML(url)
	if err != nil {
		return fmt.Errorf("failed to fetch main page: %v", err)
	}

	// Parse and extract assets
	links := extractLinks(htmlContent, url)

	// Convert rejectTypes into a map for filtering
	rejectMap := make(map[string]bool)
	if rejectTypes != "" {
		for _, ext := range strings.Split(rejectTypes, ",") {
			rejectMap[strings.TrimSpace(ext)] = true
		}
	}

	// Download extracted links
	for _, link := range links {
		if shouldReject(link, rejectMap) {
			fmt.Println("Skipping:", link)
			continue
		}

		err := downloadAsset(link, outputDir)
		if err != nil {
			fmt.Println("Error downloading:", link, err)
		}
	}

	// Save main HTML file
	mainFilePath := filepath.Join(outputDir, "index.html")
	err = os.WriteFile(mainFilePath, []byte(htmlContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to save index.html: %v", err)
	}

	endTime := time.Now()
	fmt.Printf("Website mirrored successfully! Saved in '%s'\n", outputDir)
	fmt.Printf("Completed at: %s\n", endTime.Format("2006-01-02 15:04:05"))

	return nil
}
// Fetches HTML of the main page
func fetchHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(htmlBytes), nil
}
// Extracts all relevant links from HTML
func extractLinks(html string, baseURL string) []string {
	var links []string

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return links
	}

	// Extract links from `a`, `link`, `img`, `script`, and `source` tags
	doc.Find("a, link, img, script, source").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("href")
		if !exists {
			src, exists = s.Attr("src")
		}

		if exists {
			fullURL := resolveURL(src, baseURL)
			links = append(links, fullURL)
		}
	})

	return links
}
// Converts relative URLs to absolute URLs
func resolveURL(href string, baseURL string) string {
	parsedBase, _ := url.Parse(baseURL)
	parsedHref, err := url.Parse(href)
	if err != nil {
		return ""
	}

	return parsedBase.ResolveReference(parsedHref).String()
}
// Downloads assets like images, CSS, and JS
func downloadAsset(assetURL string, outputDir string) error {
	resp, err := http.Get(assetURL)
	if err != nil {
		return fmt.Errorf("failed to fetch asset: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	fileName := filepath.Base(assetURL)
	filePath := filepath.Join(outputDir, fileName)

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	fmt.Println("Downloaded:", filePath)
	return nil
}
// Determines whether a file type should be rejected
func shouldReject(url string, rejectMap map[string]bool) bool {
	ext := strings.TrimPrefix(filepath.Ext(url), ".")
	return rejectMap[ext]
}

// Extracts the domain name from a URL
func extractDomain(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "unknown"
	}
	return parsedURL.Hostname()
}
// func mirrorWebsite(baseURL string, level int, maxDepth int, excludes []string) {
// 	if level > maxDepth {
// 		return
// 	}

// 	fmt.Printf("Mirroring: %s (Depth: %d)\n", baseURL, level)
// 	htmlContent, err := fetchHTML(baseURL)
// 	if err != nil {
// 		fmt.Printf("Failed to fetch: %s\n", baseURL)
// 		return
// 	}

// 	htmlContent = convertLinks(htmlContent, baseURL)
// 	savePath := resolveURL(baseURL)
// 	saveToFile(savePath, htmlContent)

// 	links := extractLinks(htmlContent, baseURL)
// 	for _, link := range links {
// 		if shouldExclude(link, excludes) {
// 			continue
// 		}
// 		mirrorWebsite(link, level+1, maxDepth, excludes)
// 	}
// }

// downloadFile downloads a file with speed limiting
func downloadFile(url string, outputPath string, speedLimit float64) error {
	startTime := time.Now()
	fmt.Printf("Start Time: %s\n", startTime.Format("2006-01-02 15:04:05"))

	// HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}
	fmt.Printf("Status: %s\n", resp.Status)

	// Get content length
	contentLength := resp.ContentLength
	fmt.Printf("File Size: %.2f MB (%d bytes)\n", float64(contentLength)/(1024*1024), contentLength)

	// Determine file name
	if outputPath == "" {
		outputPath = path.Base(url)
	}

	// Create file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()
	
	// Get absolute save path
	absPath, _ := filepath.Abs(outputPath)

	// Display file name and save path
	fmt.Printf("File Name: %s\n", filepath.Base(outputPath))
	fmt.Printf("Save Path: %s\n", absPath)
	

	// Apply speed limiting
	var reader io.Reader = resp.Body
	if speedLimit > 0 {
		limiter := rate.NewLimiter(rate.Limit(speedLimit*1024), int(speedLimit*1024)) // Limit in KB/s
		reader = &rateLimitedReader{reader, limiter}
	}
	
	// Progress bar variables
	buffer := make([]byte, 4096)
	var downloaded int64
	lastPercent := -5 // To ensure first update prints
	progressBarLength := 20

	// Track progress
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			if _, writeErr := outFile.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %v", writeErr)
			}
			downloaded += int64(n)

			// Calculate progress
			progress := float64(downloaded) / float64(contentLength) * 100
			elapsed := time.Since(startTime).Seconds()
			speed := float64(downloaded) / elapsed // Bytes per second
			remainingTime := float64(contentLength-downloaded) / speed // Estimated time remaining

			// Convert downloaded bytes to KiB or MiB
			var downloadedSize, totalSize float64
			var unit string
			if contentLength >= 1024*1024 {
				downloadedSize = float64(downloaded) / (1024 * 1024)
				totalSize = float64(contentLength) / (1024 * 1024)
				unit = "MiB"
			} else {
				downloadedSize = float64(downloaded) / 1024
				totalSize = float64(contentLength) / 1024
				unit = "KiB"
			}

			// Update progress every 5%
			if int(progress)-lastPercent >= 5 {
				lastPercent = int(progress)

				// Generate graphical progress bar
				progressBlocks := int(progress / (100 / float64(progressBarLength)))
				bar := "[" + strings.Repeat("=", progressBlocks) + strings.Repeat(" ", progressBarLength-progressBlocks) + "]"

				// Print progress
				fmt.Printf("\r%s %.2f%%  %.2f %s / %.2f %s  Time Remaining: %.2fs",
					bar, progress, downloadedSize, unit, totalSize, unit, remainingTime)
			}
		}

		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("error during download: %v", err)
		}
	}
	
	fmt.Println("\nDownload Complete!")

	// Show completion time
	endTime := time.Now()
	fmt.Printf("Completion Time: %s\n", endTime.Format("2006-01-02 15:04:05"))
	return nil
}

// rateLimitedReader wraps an io.Reader to enforce speed limits
type rateLimitedReader struct {
	reader  io.Reader
	limiter *rate.Limiter
}

func (r *rateLimitedReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.limiter.WaitN(nil, n)
	return n, err
}

func convertLinksFunction(filePath string) error {
	// Open the HTML file
	file, err := os.Open(filePath)
	if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Parse the HTML content
	doc, err := html.Parse(file)
	if err != nil {
			return fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Define the base URL (assuming the HTML file is in the local directory)
	base, err := url.Parse("file://" + filepath.ToSlash(filepath.Dir(filePath)) + "/")
	if err != nil {
			return fmt.Errorf("failed to parse base URL: %v", err)
	}

	// Traverse and modify the links
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
			if n.Type == html.ElementNode {
					for _, attr := range []string{"href", "src"} {
							for i, a := range n.Attr {
									if a.Key == attr {
											// Parse the URL
											u, err := url.Parse(a.Val)
											if err != nil || u.Scheme == "mailto" {
													continue
											}

											// Resolve the URL against the base
											resolved := base.ResolveReference(u)

											// Convert to relative path
											relPath, err := filepath.Rel(base.Path, resolved.Path)
											if err != nil {
													continue
											}

											// Update the attribute with the relative path
											n.Attr[i].Val = relPath
									}
							}
					}
			}
			// Recursively traverse the child nodes
			for c := n.FirstChild; c != nil; c = c.NextSibling {
					traverse(c)
			}
	}
	traverse(doc)

	// Serialize the modified HTML back to a buffer
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
			return fmt.Errorf("failed to render HTML: %v", err)
	}

	// Write the buffer to the file (or handle it as needed)
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write updated HTML to file: %v", err)
	}

	return nil
}

func main() {
	// Define command-line flags
	var mirrorFlag bool
	var convertLinks bool
	var rejectType string
	var acceptType string
	var recursive bool

	flag.BoolVar(&mirrorFlag, "mirror", false, "Mirror the remote directory structure")
	flag.BoolVar(&convertLinks, "convert-links", false, "Convert links for offline viewing")
	flag.StringVar(&rejectType, "reject", "", "Comma-separated list of file types to reject")
	flag.StringVar(&acceptType, "accept", "", "Comma-separated list of file types to accept")
	flag.BoolVar(&recursive, "recursive", false, "Download directories recursively")
	// Parse CLI arguments
	speedLimit := flag.Float64("limit", 0, "Limit download speed in KB/s (0 = unlimited)")
	outputFile := flag.String("O", "", "Rename the downloaded file")
	outputDir := flag.String("P", "", "Specify directory to save the file")
	// Define CLI flags
	// mirror := flag.Bool("mirror", false, "Mirror the remote directory structure")
	// recursive := flag.Bool("r", false, "Download files recursively")
	// rejectType := flag.String("R", "", "Comma-separated list of rejected file types")
	// acceptType := flag.String("A", "", "Comma-separated list of accepted file types")
	_ = rejectType
	_ = acceptType
	_ = recursive
	flag.Parse()
	// Ensure a URL is provided
	if flag.NArg() < 1 {
		fmt.Println("Usage: go-wget <URL> [OPTIONS]")
		os.Exit(1)
	}
	// Define `url` before using it
	url := flag.Arg(0)
	outputPath := ""
	if flag.NArg() > 1 {
		outputPath = flag.Arg(1)
	}
	// Implement the mirror functionality
	if mirrorFlag {
		err := mirrorWebsite(url, rejectType, acceptType, recursive)
		if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
		}
	}
	// Implement the convert-links functionality
	if convertLinks {
		err := convertLinksFunction(url)
		if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
		}
	}
	// Example usage
	err := convertLinksFunction("downloaded_page.html")
	if err != nil {
			fmt.Println("Error:", err)
	} else {
			fmt.Println("Links converted successfully.")
	}

	// if mirrorFlag {
	// 	err := mirrorWebsite(url, *rejectType, *acceptType, *recursive)
	// 	if err != nil {
	// 		fmt.Println("Error:", err)
	// 		os.Exit(1)
	// 	}
	// 	fmt.Println("Mirror mode enabled")
	// }

	// if mirror mode is enabled, call mirror func
	// if *mirror {
	// 	err := mirrorWebsite(url, *rejectType, *acceptType, *recursive)
	// 	if err != nil {
	// 		fmt.Println("Error:", err)
	// 		os.Exit(1)
	// 	}
	// 	return
	// }
	// else, download the file
	err = downloadFile(url, outputPath, *speedLimit)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Determine output file name
	fileName := *outputFile
	if fileName == "" {
		fileName = filepath.Base(url) // Extract filename from URL if not provided
	}
	// Determine output directory
	savePath := fileName
	if *outputDir != "" {
		expandedDir := *outputDir
		if strings.HasPrefix(expandedDir, "~") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("Error: Unable to get home directory")
				os.Exit(1)
			}
			expandedDir = filepath.Join(homeDir, expandedDir[1:])
		}

		// Create the directory if it doesn't exist
		err := os.MkdirAll(expandedDir, os.ModePerm)
		if err != nil {
			fmt.Println("Error: Unable to create directory", expandedDir)
			os.Exit(1)
		}

		savePath = filepath.Join(expandedDir, fileName)
	}

	// Call the download function
	err = downloadFile(url, savePath, *speedLimit)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
