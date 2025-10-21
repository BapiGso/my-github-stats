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

	"github.com/labstack/echo/v4"
)

// extractSVGContent returns inner content of the first <svg>...</svg>
func extractSVGContent(svg string) string {
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

func writeSVGHeadersEcho(c echo.Context) {
	h := c.Response().Header()
	h.Set("Content-Type", "image/svg+xml")
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Cache-Control", "s-maxage=3600, stale-while-revalidate")
}

func statsEchoHandler(c echo.Context) error {
	q := c.QueryParams()
	username := strings.TrimSpace(c.QueryParam("username"))
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username is required"})
	}

	// Defaults for outer SVG size
	svgWidth := 500
	svgHeight := 200
	if v := c.QueryParam("svg_width"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			svgWidth = n
		}
	}
	if v := c.QueryParam("svg_height"); v != "" {
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
	upstreamQ.Set("username", username)

	// Fetch upstream stats SVG
	orig, err := fetchUpstream("https://github-readme-stats.vercel.app/api", upstreamQ)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch upstream"})
	}
	leftContent := extractSVGContent(orig)

	// Load role asset if provided
	role := sanitizeRole(c.QueryParam("role"))
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
	if v := c.QueryParam("role_x"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			assetX = n
		}
	}
	if v := c.QueryParam("role_y"); v != "" {
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

	writeSVGHeadersEcho(c)
	return c.Blob(http.StatusOK, "image/svg+xml", []byte(sb.String()))
}

func registerStats(e *echo.Echo) {
	e.GET("/api/stats", statsEchoHandler)
}
