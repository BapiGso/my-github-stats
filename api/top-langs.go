package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func topLangsEchoHandler(c echo.Context) error {
	q := c.QueryParams()
	username := strings.TrimSpace(c.QueryParam("username"))
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username is required"})
	}

	// Defaults for outer SVG size
	svgWidth := 500
	svgHeight := 170
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

	// Prepare upstream query params (exclude our own)
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
	if upstreamQ.Get("layout") == "" {
		upstreamQ.Set("layout", "compact")
	}

	// Fetch upstream top-langs SVG
	orig, err := fetchUpstream("https://github-readme-stats.vercel.app/api/top-langs/", upstreamQ)
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

	var sb strings.Builder
	sb.Grow(len(leftContent) + len(assetContent) + 512)
	sb.WriteString(`<svg width="`)
	sb.WriteString(strconv.Itoa(svgWidth))
	sb.WriteString(`" height="`)
	sb.WriteString(strconv.Itoa(svgHeight))
	sb.WriteString(`" xmlns="http://www.w3.org/2000/svg">`)
	sb.WriteString(`\n  <rect x="0.5" y="0.5" rx="4.5" height="99%" width="`)
	sb.WriteString(strconv.Itoa(svgWidth - 1))
	sb.WriteString(`" fill="#fffefe" stroke="#e4e2e2" stroke-opacity="1"/>`)
	sb.WriteString(`\n  <g>\n`)
	sb.WriteString(leftContent)
	sb.WriteString(`\n  </g>`)
	if assetContent != "" {
		sb.WriteString(`\n  <g transform=\"translate(`)
		sb.WriteString(strconv.Itoa(assetX))
		sb.WriteString(`,`)
		sb.WriteString(strconv.Itoa(assetY))
		sb.WriteString(`)\">\n`)
		sb.WriteString(assetContent)
		sb.WriteString(`\n  </g>`)
	}
	sb.WriteString(`\n</svg>`)

	writeSVGHeadersEcho(c)
	return c.Blob(http.StatusOK, "image/svg+xml", []byte(sb.String()))
}

func registerTopLangs(e *echo.Echo) {
	e.GET("/api/top-langs", topLangsEchoHandler)
}
