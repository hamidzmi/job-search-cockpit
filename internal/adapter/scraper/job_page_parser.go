package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"job-search-cockpit/internal/domain"
)

type HTTPJobPageParser struct {
	client *http.Client
}

func NewJobPageParser(client *http.Client) *HTTPJobPageParser {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &HTTPJobPageParser{client: client}
}

var (
	jsonLDRegex       = regexp.MustCompile(`(?is)<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	scriptStyleRegex  = regexp.MustCompile(`(?is)<(script|style|noscript)[^>]*>.*?</(script|style|noscript)>`)
	liTagRegex        = regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
	blockTagRegex     = regexp.MustCompile(`(?i)</?(p|div|br|li|ul|ol|h[1-6]|tr|section|article)[^>]*>`)
	anyTagRegex       = regexp.MustCompile(`(?s)<[^>]+>`)
	multiSpaceRegex   = regexp.MustCompile(`[ \t]+`)
	multiNewlineRegex = regexp.MustCompile(`\n{3,}`)

	titleTagRegex      = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	h1TagRegex         = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	ghTitleRegex       = regexp.MustCompile(`(?is)<[^>]*class=["'][^"']*app-title[^"']*["'][^>]*>(.*?)</[^>]+>`)
	ghCompanyRegex     = regexp.MustCompile(`(?is)<[^>]*class=["'][^"']*company-name[^"']*["'][^>]*>(.*?)</[^>]+>`)
	ghLocationRegex    = regexp.MustCompile(`(?is)<[^>]*class=["'][^"']*location[^"']*["'][^>]*>(.*?)</[^>]+>`)
	leverHeadlineRegex = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*posting-headline[^"']*["'][^>]*>.*?<h2[^>]*>(.*?)</h2>`)
	ashbyTitleRegex    = regexp.MustCompile(`(?is)<[^>]+(?:data-testid=["'](?:job-title|title)["']|class=["'][^"']*(?:_headingTitle_|_title_|ashby-job-posting-heading|ashby-job-posting-title)[^"']*["'])[^>]*>(.*?)</[^>]+>`)
	ashbyLocationRegex = regexp.MustCompile(`(?is)<[^>]+(?:data-testid=["'](?:job-location|location)["']|class=["'][^"']*(?:_location_|_jobPostingBrief_|ashby-job-posting-brief|ashby-job-posting-location)[^"']*["'])[^>]*>(.*?)</[^>]+>`)
	ashbyCompanyRegex  = regexp.MustCompile(`(?is)<[^>]+(?:data-testid=["']company-name["']|class=["'][^"']*(?:_companyName_|ashby-job-posting-company)[^"']*["'])[^>]*>(.*?)</[^>]+>`)
)

func (p *HTTPJobPageParser) ParseJobPage(ctx context.Context, rawURL string) (domain.ParsedJobPage, error) {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return domain.ParsedJobPage{}, fmt.Errorf("%w: empty job posting url", domain.ErrValidation)
	}
	parsedURL, err := url.Parse(trimmedURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return domain.ParsedJobPage{}, fmt.Errorf("%w: invalid job posting url %q", domain.ErrValidation, trimmedURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, trimmedURL, nil)
	if err != nil {
		return domain.ParsedJobPage{}, fmt.Errorf("create page request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; JobSearchCockpit/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := p.client.Do(req)
	if err != nil {
		return domain.ParsedJobPage{}, fmt.Errorf("fetch %s: %w", trimmedURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.ParsedJobPage{}, fmt.Errorf("fetch %s returned HTTP %d", trimmedURL, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return domain.ParsedJobPage{}, fmt.Errorf("read %s body: %w", trimmedURL, err)
	}

	return ParseJobPageHTML(trimmedURL, string(bodyBytes)), nil
}

func ParseJobPageHTML(sourceURL, rawHTML string) domain.ParsedJobPage {
	out := domain.ParsedJobPage{
		SourceURL:      strings.TrimSpace(sourceURL),
		DetectedSource: "html",
	}

	// 1. Try JSON-LD JobPosting schema
	if ldPage, ok := extractJSONLDJobPosting(rawHTML); ok {
		out.CompanyName = ldPage.CompanyName
		out.CompanyDomain = ldPage.CompanyDomain
		out.RoleTitle = ldPage.RoleTitle
		out.Seniority = ldPage.Seniority
		out.Location = ldPage.Location
		out.RemotePolicy = ldPage.RemotePolicy
		out.StatedMinComp = ldPage.StatedMinComp
		out.StatedMaxComp = ldPage.StatedMaxComp
		out.Currency = ldPage.Currency
		out.Requirements = ldPage.Requirements
		out.TechStackTags = ldPage.TechStackTags
		out.CleanedText = ldPage.CleanedText
		out.DetectedSource = "json-ld"
	}

	// 2. Enrich from ATS-specific layouts (Greenhouse, Lever, Ashby) and HTML structure
	enrichFromATSHTML(&out, sourceURL, rawHTML)

	cleanedPageText := cleanHTMLToText(rawHTML)
	if out.CleanedText == "" {
		out.CleanedText = cleanedPageText
	}

	if len(out.Requirements) == 0 {
		out.Requirements = extractListItems(rawHTML)
	}

	combinedText := out.CleanedText + "\n" + cleanedPageText
	if out.StatedMinComp == 0 && out.StatedMaxComp == 0 {
		minC, maxC := domain.ExtractCompensationRange(combinedText)
		if minC > 0 || maxC > 0 {
			out.StatedMinComp = minC
			out.StatedMaxComp = maxC
			if out.Currency == "" {
				out.Currency = "USD"
			}
		}
	}
	if out.RemotePolicy == "" {
		out.RemotePolicy = domain.InferRemotePolicy(out.Location + " " + combinedText)
	}
	if out.Seniority == "" && out.RoleTitle != "" {
		out.Seniority = domain.InferSeniorityFromTitle(out.RoleTitle)
	}
	if len(out.TechStackTags) == 0 {
		out.TechStackTags = domain.ExtractKnownTechTags(combinedText)
	}

	return out
}

func extractJSONLDJobPosting(rawHTML string) (domain.ParsedJobPage, bool) {
	matches := jsonLDRegex.FindAllStringSubmatch(rawHTML, -1)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		scriptContent := strings.TrimSpace(html.UnescapeString(m[1]))
		var generic any
		if err := json.Unmarshal([]byte(scriptContent), &generic); err != nil {
			continue
		}
		if obj := findJobPostingNode(generic); obj != nil {
			return parseJobPostingNode(obj), true
		}
	}
	return domain.ParsedJobPage{}, false
}

func findJobPostingNode(node any) map[string]any {
	switch v := node.(type) {
	case map[string]any:
		if isJobPostingType(v["@type"]) {
			return v
		}
		if graph, ok := v["@graph"].([]any); ok {
			for _, item := range graph {
				if found := findJobPostingNode(item); found != nil {
					return found
				}
			}
		}
	case []any:
		for _, item := range v {
			if found := findJobPostingNode(item); found != nil {
				return found
			}
		}
	}
	return nil
}

func isJobPostingType(t any) bool {
	switch v := t.(type) {
	case string:
		return strings.EqualFold(strings.TrimSpace(v), "JobPosting")
	case []any:
		for _, elem := range v {
			if s, ok := elem.(string); ok && strings.EqualFold(strings.TrimSpace(s), "JobPosting") {
				return true
			}
		}
	}
	return false
}

func parseJobPostingNode(node map[string]any) domain.ParsedJobPage {
	var page domain.ParsedJobPage
	page.RoleTitle = asString(node["title"])
	page.Seniority = domain.InferSeniorityFromTitle(page.RoleTitle)

	if org, ok := node["hiringOrganization"].(map[string]any); ok {
		page.CompanyName = asString(org["name"])
		sameAs := asString(org["sameAs"])
		if sameAs != "" {
			if u, err := url.Parse(sameAs); err == nil && u.Host != "" {
				page.CompanyDomain = strings.TrimPrefix(u.Host, "www.")
			} else {
				page.CompanyDomain = sameAs
			}
		}
	}

	page.Location = extractJSONLDLocation(node["jobLocation"])
	locType := strings.ToUpper(asString(node["jobLocationType"]))
	if strings.Contains(locType, "TELECOMMUTE") || strings.Contains(locType, "REMOTE") {
		page.RemotePolicy = domain.RemotePolicyRemote
	}

	if salary, ok := node["baseSalary"].(map[string]any); ok {
		page.Currency = asString(salary["currency"])
		if valMap, ok := salary["value"].(map[string]any); ok {
			page.StatedMinComp = asInt(valMap["minValue"])
			page.StatedMaxComp = asInt(valMap["maxValue"])
			if page.StatedMinComp == 0 && page.StatedMaxComp == 0 {
				v := asInt(valMap["value"])
				page.StatedMinComp = v
				page.StatedMaxComp = v
			}
			if page.Currency == "" {
				page.Currency = asString(valMap["currency"])
			}
		} else {
			v := asInt(salary["value"])
			page.StatedMinComp = v
			page.StatedMaxComp = v
		}
	}

	descHTML := asString(node["description"])
	if descHTML != "" {
		page.CleanedText = cleanHTMLToText(descHTML)
		page.Requirements = extractListItems(descHTML)
		page.TechStackTags = domain.ExtractKnownTechTags(page.CleanedText)
	}

	return page
}

func extractJSONLDLocation(locNode any) string {
	switch v := locNode.(type) {
	case map[string]any:
		if addr, ok := v["address"].(map[string]any); ok {
			locality := asString(addr["addressLocality"])
			region := asString(addr["addressRegion"])
			country := asString(addr["addressCountry"])
			parts := make([]string, 0, 3)
			for _, p := range []string{locality, region, country} {
				if p != "" {
					parts = append(parts, p)
				}
			}
			return strings.Join(parts, ", ")
		}
		return asString(v["name"])
	case []any:
		for _, item := range v {
			if s := extractJSONLDLocation(item); s != "" {
				return s
			}
		}
	}
	return ""
}

func enrichFromATSHTML(out *domain.ParsedJobPage, sourceURL, rawHTML string) {
	lowerURL := strings.ToLower(sourceURL)
	lowerHTML := strings.ToLower(rawHTML)

	// Greenhouse layout
	if strings.Contains(lowerURL, "greenhouse.io") || strings.Contains(lowerHTML, "app-title") {
		if out.DetectedSource == "html" {
			out.DetectedSource = "greenhouse"
		}
		if out.RoleTitle == "" {
			out.RoleTitle = extractFirstGroupText(ghTitleRegex, rawHTML)
		}
		if out.CompanyName == "" {
			comp := extractFirstGroupText(ghCompanyRegex, rawHTML)
			comp = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(comp), "at "))
			out.CompanyName = comp
		}
		if out.Location == "" {
			out.Location = extractFirstGroupText(ghLocationRegex, rawHTML)
		}
	}

	// Lever layout
	if strings.Contains(lowerURL, "lever.co") || strings.Contains(lowerHTML, "posting-headline") {
		if out.DetectedSource == "html" {
			out.DetectedSource = "lever"
		}
		if out.RoleTitle == "" {
			out.RoleTitle = extractFirstGroupText(leverHeadlineRegex, rawHTML)
		}
		if out.Location == "" {
			out.Location = extractFirstGroupText(ghLocationRegex, rawHTML)
		}
	}

	// Ashby layout
	if strings.Contains(lowerURL, "ashbyhq.com") || strings.Contains(lowerHTML, "ashby-job-posting") || strings.Contains(lowerHTML, "jobs.ashbyhq") {
		if out.DetectedSource == "html" {
			out.DetectedSource = "ashby"
		}
		if out.RoleTitle == "" {
			out.RoleTitle = extractFirstGroupText(ashbyTitleRegex, rawHTML)
		}
		if out.RoleTitle == "" {
			out.RoleTitle = extractMetaTag(rawHTML, "og:title")
		}
		if out.RoleTitle == "" {
			out.RoleTitle = extractMetaTag(rawHTML, "twitter:title")
		}
		if out.CompanyName == "" {
			comp := extractFirstGroupText(ashbyCompanyRegex, rawHTML)
			if comp == "" {
				comp = extractMetaTag(rawHTML, "og:site_name")
			}
			if comp == "" {
				comp = extractAshbyCompanyFromURL(sourceURL)
			}
			out.CompanyName = comp
		}
		if out.Location == "" {
			loc := extractFirstGroupText(ashbyLocationRegex, rawHTML)
			if loc == "" {
				loc = extractMetaTag(rawHTML, "ashby:location")
			}
			if loc == "" {
				loc = extractMetaTag(rawHTML, "location")
			}
			if loc == "" {
				loc = extractMetaTag(rawHTML, "og:locality")
			}
			out.Location = loc
		}
	}

	// Title / H1 / OpenGraph meta fallback
	titleText := extractFirstGroupText(titleTagRegex, rawHTML)
	if titleText == "" {
		titleText = extractMetaTag(rawHTML, "og:title")
	}
	if out.RoleTitle == "" {
		out.RoleTitle = extractFirstGroupText(h1TagRegex, rawHTML)
	}
	if titleText != "" && (out.RoleTitle == "" || out.CompanyName == "") {
		parseTitleTagForRoleAndCompany(titleText, out)
	}
	if out.CompanyName == "" {
		out.CompanyName = extractMetaTag(rawHTML, "og:site_name")
	}
	if out.Location == "" {
		out.Location = extractMetaTag(rawHTML, "location")
	}
}

func extractMetaTag(rawHTML, nameOrProp string) string {
	pattern := fmt.Sprintf(`(?is)<meta[^>]+(?:property|name)=["']%s["'][^>]+content=["']([^"']*)["']`, regexp.QuoteMeta(nameOrProp))
	re := regexp.MustCompile(pattern)
	if m := re.FindStringSubmatch(rawHTML); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
		return cleanInlineHTML(m[1])
	}
	patternRev := fmt.Sprintf(`(?is)<meta[^>]+content=["']([^"']*)["'][^>]+(?:property|name)=["']%s["']`, regexp.QuoteMeta(nameOrProp))
	reRev := regexp.MustCompile(patternRev)
	if m := reRev.FindStringSubmatch(rawHTML); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
		return cleanInlineHTML(m[1])
	}
	return ""
}

func extractAshbyCompanyFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) > 0 && parts[0] != "" {
		slug := parts[0]
		words := strings.Fields(strings.ReplaceAll(slug, "-", " "))
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
		}
		return strings.Join(words, " ")
	}
	return ""
}

func parseTitleTagForRoleAndCompany(titleText string, out *domain.ParsedJobPage) {
	cleaned := strings.TrimSpace(titleText)
	lower := strings.ToLower(cleaned)
	if strings.HasPrefix(lower, "job application for ") {
		cleaned = strings.TrimSpace(cleaned[len("job application for "):])
	}
	if idx := strings.LastIndex(strings.ToLower(cleaned), " at "); idx != -1 {
		role := strings.TrimSpace(cleaned[:idx])
		comp := strings.TrimSpace(cleaned[idx+4:])
		if out.RoleTitle == "" {
			out.RoleTitle = role
		}
		if out.CompanyName == "" {
			out.CompanyName = comp
		}
		return
	}
	for _, sep := range []string{" - ", " | ", " @ "} {
		if parts := strings.Split(cleaned, sep); len(parts) >= 2 {
			first := strings.TrimSpace(parts[0])
			last := strings.TrimSpace(parts[len(parts)-1])
			if out.RoleTitle == "" {
				out.RoleTitle = first
			}
			if out.CompanyName == "" {
				if strings.EqualFold(out.RoleTitle, last) {
					out.CompanyName = first
				} else {
					out.CompanyName = last
				}
			}
			return
		}
	}
	if out.RoleTitle == "" {
		out.RoleTitle = cleaned
	}
}

func extractFirstGroupText(re *regexp.Regexp, rawHTML string) string {
	m := re.FindStringSubmatch(rawHTML)
	if len(m) < 2 {
		return ""
	}
	return cleanInlineHTML(m[1])
}

func extractListItems(rawHTML string) []string {
	matches := liTagRegex.FindAllStringSubmatch(rawHTML, 10)
	items := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		text := cleanInlineHTML(m[1])
		if text != "" {
			items = append(items, text)
		}
	}
	return items
}

func cleanInlineHTML(fragment string) string {
	noTags := anyTagRegex.ReplaceAllString(fragment, " ")
	unescaped := html.UnescapeString(noTags)
	return strings.TrimSpace(multiSpaceRegex.ReplaceAllString(unescaped, " "))
}

func cleanHTMLToText(rawHTML string) string {
	withoutScripts := scriptStyleRegex.ReplaceAllString(rawHTML, " ")
	withNewlines := blockTagRegex.ReplaceAllString(withoutScripts, "\n")
	withoutTags := anyTagRegex.ReplaceAllString(withNewlines, " ")
	unescaped := html.UnescapeString(withoutTags)

	lines := strings.Split(unescaped, "\n")
	cleanedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(multiSpaceRegex.ReplaceAllString(line, " "))
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}
	joined := strings.Join(cleanedLines, "\n")
	return strings.TrimSpace(multiNewlineRegex.ReplaceAllString(joined, "\n\n"))
}

func asString(v any) string {
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	default:
		return ""
	}
}

func asInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		cleaned := strings.ReplaceAll(strings.TrimSpace(n), ",", "")
		if f, err := strconv.ParseFloat(cleaned, 64); err == nil {
			return int(f)
		}
	}
	return 0
}
