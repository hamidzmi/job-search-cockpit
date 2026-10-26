package domain

import (
	"regexp"
	"strconv"
	"strings"
)

type ParsedJobPage struct {
	SourceURL      string
	CompanyName    string
	CompanyDomain  string
	RoleTitle      string
	Seniority      string
	Location       string
	RemotePolicy   RemotePolicy
	StatedMinComp  int
	StatedMaxComp  int
	Currency       string
	TechStackTags  []string
	Requirements   []string
	CleanedText    string
	DetectedSource string
}

type ExtractedJobPostingDraft struct {
	CompanyName      string
	CompanyDomain    string
	RoleTitle        string
	Seniority        string
	Location         string
	RemotePolicy     RemotePolicy
	StatedMinComp    int
	StatedMaxComp    int
	Currency         string
	TechStackTags    []string
	Requirements     []string
	SourceURL        string
	RawText          string
	ExtractionSource string
	ModelUsed        string
	FetchWarning     string
	AssistantWarning string
}

type QuickCaptureExtractInput struct {
	SourceURL     string
	RawText       string
	ModelOverride string
}

type AssistantExtractInput struct {
	SourceURL         string
	RawText           string
	DeterministicHint ParsedJobPage
	Model             string
}

type QuickCaptureSaveInput struct {
	CompanyID       string
	CompanyName     string
	CompanyDomain   string
	RoleTitle       string
	Seniority       string
	Location        string
	RemotePolicy    RemotePolicy
	Priority        Priority
	Stage           Stage
	TransitionNote  string
	Notes           string
	JobPosting      JobPosting
}

var (
	compRangeRegex = regexp.MustCompile(`(?i)\$\s*([0-9]{2,3}(?:,[0-9]{3})+|([0-9]{2,3})k)\s*(?:-|–|—|to)\s*\$?\s*([0-9]{2,3}(?:,[0-9]{3})+|([0-9]{2,3})k)`)
	knownTechTags  = []string{
		"Go", "Golang", "Rust", "Python", "TypeScript", "JavaScript", "React",
		"Ruby", "Sorbet", "Java", "Kotlin", "C++", "PostgreSQL", "SQLite",
		"MySQL", "Redis", "Kafka", "GraphQL", "gRPC", "Docker", "Kubernetes",
		"AWS", "GCP", "Terraform", "eBPF",
	}
)

func MergeCaptureDraft(sourceURL, userRawText string, page ParsedJobPage, llm *ExtractedJobPostingDraft) ExtractedJobPostingDraft {
	trimmedURL := strings.TrimSpace(sourceURL)
	if trimmedURL == "" {
		trimmedURL = strings.TrimSpace(page.SourceURL)
	}
	rawText := strings.TrimSpace(userRawText)
	if rawText == "" {
		rawText = strings.TrimSpace(page.CleanedText)
	}

	heuristics := ExtractHeuristicsFromText(rawText)

	draft := ExtractedJobPostingDraft{
		CompanyName:      firstNonEmpty(page.CompanyName, heuristics.CompanyName),
		CompanyDomain:    strings.TrimSpace(page.CompanyDomain),
		RoleTitle:        firstNonEmpty(page.RoleTitle, heuristics.RoleTitle),
		Seniority:        firstNonEmpty(page.Seniority, heuristics.Seniority),
		Location:         firstNonEmpty(page.Location, heuristics.Location),
		RemotePolicy:     firstRemotePolicy(page.RemotePolicy, heuristics.RemotePolicy),
		StatedMinComp:    firstPositiveInt(page.StatedMinComp, heuristics.StatedMinComp),
		StatedMaxComp:    firstPositiveInt(page.StatedMaxComp, heuristics.StatedMaxComp),
		Currency:         firstNonEmpty(page.Currency, heuristics.Currency, "USD"),
		TechStackTags:    dedupStrings(append(page.TechStackTags, heuristics.TechStackTags...)),
		Requirements:     firstNonEmptySlice(page.Requirements, heuristics.Requirements),
		SourceURL:        trimmedURL,
		RawText:          rawText,
		ExtractionSource: "manual_fallback",
	}

	if hasDeterministicPageFields(page) {
		draft.ExtractionSource = "deterministic"
	}

	if draft.Seniority == "" && draft.RoleTitle != "" {
		draft.Seniority = InferSeniorityFromTitle(draft.RoleTitle)
	}

	if llm != nil {
		draft.CompanyName = firstNonEmpty(llm.CompanyName, draft.CompanyName)
		draft.CompanyDomain = firstNonEmpty(llm.CompanyDomain, draft.CompanyDomain)
		draft.RoleTitle = firstNonEmpty(llm.RoleTitle, draft.RoleTitle)
		draft.Seniority = firstNonEmpty(llm.Seniority, draft.Seniority)
		if draft.Seniority == "" && draft.RoleTitle != "" {
			draft.Seniority = InferSeniorityFromTitle(draft.RoleTitle)
		}
		draft.Location = firstNonEmpty(llm.Location, draft.Location)
		draft.RemotePolicy = firstRemotePolicy(llm.RemotePolicy, draft.RemotePolicy)
		draft.StatedMinComp = firstPositiveInt(llm.StatedMinComp, draft.StatedMinComp)
		draft.StatedMaxComp = firstPositiveInt(llm.StatedMaxComp, draft.StatedMaxComp)
		draft.Currency = firstNonEmpty(llm.Currency, draft.Currency, "USD")
		draft.TechStackTags = firstNonEmptySlice(dedupStrings(llm.TechStackTags), draft.TechStackTags)
		draft.Requirements = firstNonEmptySlice(dedupStrings(llm.Requirements), draft.Requirements)
		draft.ExtractionSource = "llm"
		draft.ModelUsed = strings.TrimSpace(llm.ModelUsed)
	}

	return draft
}

func ExtractHeuristicsFromText(text string) ParsedJobPage {
	var out ParsedJobPage
	if strings.TrimSpace(text) == "" {
		return out
	}

	lines := strings.Split(text, "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "company:") && out.CompanyName == "":
			out.CompanyName = strings.TrimSpace(line[len("company:"):])
		case (strings.HasPrefix(lower, "role:") || strings.HasPrefix(lower, "title:") || strings.HasPrefix(lower, "position:")) && out.RoleTitle == "":
			idx := strings.Index(line, ":")
			out.RoleTitle = strings.TrimSpace(line[idx+1:])
		case strings.HasPrefix(lower, "location:") && out.Location == "":
			out.Location = strings.TrimSpace(line[len("location:"):])
		case strings.HasPrefix(lower, "seniority:") && out.Seniority == "":
			out.Seniority = strings.TrimSpace(line[len("seniority:"):])
		case strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "• "):
			req := strings.TrimSpace(strings.TrimLeft(line, "-*• "))
			if req != "" && len(out.Requirements) < 8 {
				out.Requirements = append(out.Requirements, req)
			}
		}
	}

	if out.Seniority == "" && out.RoleTitle != "" {
		out.Seniority = InferSeniorityFromTitle(out.RoleTitle)
	}

	out.RemotePolicy = InferRemotePolicy(text + " " + out.Location)
	minC, maxC := ExtractCompensationRange(text)
	if minC > 0 || maxC > 0 {
		out.StatedMinComp = minC
		out.StatedMaxComp = maxC
		out.Currency = "USD"
	}
	out.TechStackTags = ExtractKnownTechTags(text)
	return out
}

func InferSeniorityFromTitle(title string) string {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "principal"):
		return "Principal"
	case strings.Contains(lower, "distinguished"):
		return "Distinguished"
	case strings.Contains(lower, "staff"):
		return "Staff"
	case strings.Contains(lower, "senior") || strings.Contains(lower, "sr.") || strings.Contains(lower, "sr "):
		return "Senior"
	case strings.Contains(lower, "lead"):
		return "Lead"
	case strings.Contains(lower, "founding"):
		return "Founding"
	case strings.Contains(lower, "junior") || strings.Contains(lower, "jr"):
		return "Junior"
	default:
		return ""
	}
}

func InferRemotePolicy(text string) RemotePolicy {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "hybrid"):
		return RemotePolicyHybrid
	case strings.Contains(lower, "remote") || strings.Contains(lower, "telecommute") || strings.Contains(lower, "work from anywhere"):
		return RemotePolicyRemote
	case strings.Contains(lower, "on-site") || strings.Contains(lower, "onsite") || strings.Contains(lower, "in-office"):
		return RemotePolicyOnsite
	default:
		return ""
	}
}

func ExtractCompensationRange(text string) (int, int) {
	m := compRangeRegex.FindStringSubmatch(text)
	if len(m) == 0 {
		return 0, 0
	}
	minVal := parseSalaryToken(m[1], m[2])
	maxVal := parseSalaryToken(m[3], m[4])
	if minVal > 0 && maxVal > 0 && minVal <= maxVal {
		return minVal, maxVal
	}
	return 0, 0
}

func parseSalaryToken(fullToken, kToken string) int {
	if kToken != "" {
		v, err := strconv.Atoi(kToken)
		if err == nil {
			return v * 1000
		}
	}
	cleaned := strings.ReplaceAll(strings.TrimSpace(fullToken), ",", "")
	cleaned = strings.TrimSuffix(strings.ToLower(cleaned), "k")
	v, err := strconv.Atoi(cleaned)
	if err != nil {
		return 0
	}
	return v
}

func ExtractKnownTechTags(text string) []string {
	var found []string
	lower := strings.ToLower(text)
	for _, tag := range knownTechTags {
		tagLower := strings.ToLower(tag)
		if containsWord(lower, tagLower) {
			canonical := tag
			if canonical == "Golang" {
				canonical = "Go"
			}
			found = append(found, canonical)
		}
	}
	return dedupStrings(found)
}

func containsWord(haystackLower, wordLower string) bool {
	idx := strings.Index(haystackLower, wordLower)
	for idx != -1 {
		beforeOK := idx == 0 || !isWordChar(haystackLower[idx-1])
		end := idx + len(wordLower)
		afterOK := end == len(haystackLower) || !isWordChar(haystackLower[end])
		if beforeOK && afterOK {
			return true
		}
		next := strings.Index(haystackLower[end:], wordLower)
		if next == -1 {
			break
		}
		idx = end + next
	}
	return false
}

func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

func hasDeterministicPageFields(page ParsedJobPage) bool {
	return strings.TrimSpace(page.CompanyName) != "" ||
		strings.TrimSpace(page.RoleTitle) != "" ||
		strings.TrimSpace(page.Location) != "" ||
		page.StatedMinComp > 0 ||
		page.StatedMaxComp > 0 ||
		len(page.Requirements) > 0 ||
		len(page.TechStackTags) > 0
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstRemotePolicy(values ...RemotePolicy) RemotePolicy {
	for _, v := range values {
		trimmed := RemotePolicy(strings.TrimSpace(string(v)))
		if trimmed.IsValid() {
			return trimmed
		}
	}
	return ""
}

func firstPositiveInt(values ...int) int {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}

func firstNonEmptySlice(slices ...[]string) []string {
	for _, s := range slices {
		cleaned := dedupStrings(s)
		if len(cleaned) > 0 {
			return cleaned
		}
	}
	return []string{}
}

func dedupStrings(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, s := range in {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
