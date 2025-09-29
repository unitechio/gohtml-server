// server/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type PageParameters struct {
	PaperWidth   string  `json:"paperWidth"`
	PaperHeight  string  `json:"paperHeight"`
	PageSize     *string `json:"pageSize"`
	Orientation  string  `json:"orientation"`
	MarginTop    string  `json:"marginTop"`
	MarginBottom string  `json:"marginBottom"`
	MarginLeft   string  `json:"marginLeft"`
	MarginRight  string  `json:"marginRight"`
}

type BySelector struct {
	Selector string `json:"selector"`
	By       string `json:"by"`
}

type RenderParameters struct {
	WaitTime    int64        `json:"waitTime"`
	WaitReady   []BySelector `json:"waitReady"`
	WaitVisible []BySelector `json:"waitVisible"`
}

type generatePDFRequestV1 struct {
	Content          []byte           `json:"content"`
	ContentType      string           `json:"contentType"`
	ContentURL       string           `json:"contentURL"`
	Method           string           `json:"method"`
	TimeoutDuration  int64            `json:"timeoutDuration,omitempty"`
	PageParameters   PageParameters   `json:"PageParameters"`
	RenderParameters RenderParameters `json:"RenderParameters"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/v1/pdf", pdfHandler)

	log.Println("UniHTML Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func pdfHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received PDF conversion request")

	var req generatePDFRequestV1
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pdfData, err := convertToPDF(&req)
	if err != nil {
		log.Printf("Error converting to PDF: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Job-ID", fmt.Sprintf("job-%d", time.Now().Unix()))
	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusCreated)
	w.Write(pdfData)

	log.Printf("Successfully generated PDF (%d bytes)", len(pdfData))
}

func getChromePath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Cannot get executable path: %v", err)
		return ""
	}
	exeDir := filepath.Dir(exePath)

	var chromePath string
	switch runtime.GOOS {
	case "windows":
		cwd, _ := os.Getwd()
		chromePath = filepath.Join(cwd, "bin", "chrome.exe")
	case "darwin":
		chromePath = filepath.Join(exeDir, "bin", "chrome-mac", "Chromium.app", "Contents", "MacOS", "Chromium")
	case "linux":
		chromePath = filepath.Join(exeDir, "bin", "chrome")
	default:
		log.Printf("Unsupported OS: %s", runtime.GOOS)
		return ""
	}

	// Kiểm tra file có tồn tại không
	if _, err := os.Stat(chromePath); os.IsNotExist(err) {
		log.Printf("Chrome binary not found at: %s", chromePath)
		return ""
	}

	log.Printf("Using Chrome binary at: %s", chromePath)
	return chromePath
}

func convertToPDF(req *generatePDFRequestV1) ([]byte, error) {
	timeout := 30 * time.Second
	if req.TimeoutDuration > 0 {
		timeout = time.Duration(req.TimeoutDuration) * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// Xử lý HTML content
	var htmlURL string
	if req.Method == "web" {
		htmlURL = req.ContentURL
	} else {
		tmpHTML, err := os.CreateTemp("", "unihtml-*.html")
		if err != nil {
			return nil, fmt.Errorf("create temp HTML: %w", err)
		}
		defer os.Remove(tmpHTML.Name())

		if _, err := tmpHTML.Write(req.Content); err != nil {
			return nil, fmt.Errorf("write HTML: %w", err)
		}
		tmpHTML.Close()

		htmlURL = "file://" + tmpHTML.Name()
	}

	// Parse parameters
	marginTop := parseMargin(req.PageParameters.MarginTop, 10)
	marginBottom := parseMargin(req.PageParameters.MarginBottom, 10)
	marginLeft := parseMargin(req.PageParameters.MarginLeft, 10)
	marginRight := parseMargin(req.PageParameters.MarginRight, 10)

	paperWidth := parsePaperSize(req.PageParameters.PaperWidth, 8.5)
	paperHeight := parsePaperSize(req.PageParameters.PaperHeight, 11)

	landscape := req.PageParameters.Orientation == "landscape"

	pdfParams := page.PrintToPDF().
		WithPaperWidth(paperWidth).
		WithPaperHeight(paperHeight).
		WithMarginTop(marginTop).
		WithMarginBottom(marginBottom).
		WithMarginLeft(marginLeft).
		WithMarginRight(marginRight).
		WithLandscape(landscape).
		WithPrintBackground(true)

	var pdfBuffer []byte
	var tasks chromedp.Tasks

	tasks = append(tasks, chromedp.Navigate(htmlURL))
	tasks = append(tasks, chromedp.WaitReady("body", chromedp.ByQuery))

	if req.RenderParameters.WaitTime > 0 {
		tasks = append(tasks, chromedp.Sleep(time.Duration(req.RenderParameters.WaitTime)*time.Millisecond))
	}

	for _, sel := range req.RenderParameters.WaitReady {
		byOption := getByOption(sel.By)
		tasks = append(tasks, chromedp.WaitReady(sel.Selector, byOption))
	}

	for _, sel := range req.RenderParameters.WaitVisible {
		byOption := getByOption(sel.By)
		tasks = append(tasks, chromedp.WaitVisible(sel.Selector, byOption))
	}

	tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		pdfBuffer, _, err = pdfParams.Do(ctx)
		return err
	}))

	if err := chromedp.Run(ctx, tasks); err != nil {
		return nil, fmt.Errorf("chromedp run: %w", err)
	}

	return pdfBuffer, nil
}

// Helper functions (parseMargin, parsePaperSize, getByOption) giữ nguyên như trước
func parseMargin(margin string, defaultVal float64) float64 {
	if margin == "" {
		return defaultVal / 25.4
	}
	var val float64
	fmt.Sscanf(margin, "%fmm", &val)
	if val == 0 {
		fmt.Sscanf(margin, "%f", &val)
	}
	return val / 25.4
}

func parsePaperSize(size string, defaultVal float64) float64 {
	if size == "" {
		return defaultVal
	}
	var val float64
	fmt.Sscanf(size, "%fmm", &val)
	if val == 0 {
		fmt.Sscanf(size, "%f", &val)
		return val
	}
	return val / 25.4
}

func getByOption(by string) chromedp.QueryOption {
	switch by {
	case "id":
		return chromedp.ByID
	case "xpath":
		return chromedp.BySearch
	case "css":
		return chromedp.ByQuery
	default:
		return chromedp.ByQuery
	}
}
