package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// extractSVGContent returns inner content of the first <svg>...</svg>
func extractSVGContent(svg string) string {
	// Fast path using string indexes
	startTagIdx := strings.Index(strings.ToLower(svg), "<svg")
	if startTagIdx == -1 {
		return svg
	}
	gtIdx := strings.Index(svg[startTagIdx:], ">")
	if gtIdx == -1 {
		return svg
	}
	contentStart := startTagIdx + gtIdx + 1
	endIdx := strings.LastIndex(strings.ToLower(svg), "</svg>")
	if endIdx == -1 || endIdx <= contentStart {
		return svg
	}
	return svg[contentStart:endIdx]
}

var safeNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func sanitizeRole(v string) string {
	if safeNameRe.MatchString(v) {
		return v
	}
	return ""
}

func getAssetsDir() string {
	if v := os.Getenv("ASSETS_DIR"); v != "" {
		return v
	}
	// default to ./assets relative to working directory
	return "assets"
}

func fetchUpstream(base string, q url.Values) (string, error) {
	u, _ := url.Parse(base)
	u.RawQuery = q.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeSVGHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "s-maxage=3600, stale-while-revalidate")
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	username := strings.TrimSpace(q.Get("username"))
	if username == "" {
		http.Error(w, `{"error":"Username is required"}`, http.StatusBadRequest)
		return
	}

	// Defaults for outer SVG size
	svgWidth := 500
	svgHeight := 200
	if v := q.Get("svg_width"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			svgWidth = n
		}
	}
	if v := q.Get("svg_height"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			svgHeight = n
		}
	}

	// Prepare upstream query params (exclude our own custom ones)
	upstreamQ := url.Values{}
	for key, vals := range q {
		if key == "svg_width" || key == "svg_height" || key == "role" || key == "role_x" || key == "role_y" {
			continue
		}
		for _, v := range vals {
			upstreamQ.Add(key, v)
		}
	}
	// Ensure username is set
	upstreamQ.Set("username", username)

	// Fetch upstream stats SVG
	orig, err := fetchUpstream("https://github-readme-stats.vercel.app/api", upstreamQ)
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch upstream"}`, http.StatusBadGateway)
		return
	}

	leftContent := extractSVGContent(orig)

	// Load role asset if provided
	role := sanitizeRole(q.Get("role"))
	assetContent := ""
	if role != "" {
		assetsDir := getAssetsDir()
		p := filepath.Join(assetsDir, role+".svg")
		if f, err := os.Open(p); err == nil {
			defer f.Close()
			b, err := io.ReadAll(f)
			if err == nil {
				assetContent = extractSVGContent(string(b))
			}
		}
	}

	// Place asset on the right side with a default offset
	assetX := 360
	assetY := 0
	if v := q.Get("role_x"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			assetX = n
		}
	}
	if v := q.Get("role_y"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			assetY = n
		}
	}

	// Compose final SVG
	var sb strings.Builder
	sb.Grow(len(leftContent) + len(assetContent) + 512)
	sb.WriteString(`<svg width="`)
	sb.WriteString(strconv.Itoa(svgWidth))
	sb.WriteString(`" height="`)
	sb.WriteString(strconv.Itoa(svgHeight))
	sb.WriteString(`" xmlns="http://www.w3.org/2000/svg">`)
	sb.WriteString(`\n  <rect x="0.5" y="0.5" rx="4.5" height="98%" width="`)
	sb.WriteString(strconv.Itoa(svgWidth - 1))
	sb.WriteString(`" fill="#fffefe" stroke="#e4e2e2" stroke-opacity="1"/>`)
	sb.WriteString(`\n  <g>\n`)
	sb.WriteString(leftContent)
	sb.WriteString(`\n  </g>`)
	if assetContent != "" {
		sb.WriteString(`\n  <g transform="translate(`)
		sb.WriteString(strconv.Itoa(assetX))
		sb.WriteString(`,`)
		sb.WriteString(strconv.Itoa(assetY))
		sb.WriteString(`)">\n`)
		sb.WriteString(assetContent)
		sb.WriteString(`\n  </g>`)
	}
	sb.WriteString(`\n</svg>`)

	writeSVGHeaders(w)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(sb.String()))
}

func main() {
	http.HandleFunc("/api/stats", statsHandler)
	// For local dev: allow running `go run api/stats.go` to serve this one endpoint
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	// If running on Vercel, the platform will route requests directly to the handler.
	// Running a local server here won't affect Vercel, but helps for local testing.
	if os.Getenv("GO_LOCAL_SERVER") == "1" {
		_ = http.ListenAndServe(":"+port, nil)
	}
}
