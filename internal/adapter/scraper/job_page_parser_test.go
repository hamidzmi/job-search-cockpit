package scraper_test

import (
	"testing"

	"job-search-cockpit/internal/adapter/scraper"
)

func TestParseJobPageHTML_AshbyExtraction(t *testing.T) {
	t.Run("extracts title, company, and location from Ashby HTML layout classes", func(t *testing.T) {
		rawHTML := `<!DOCTYPE html>
<html>
<head>
	<title>Staff Backend Engineer at Acme Corp</title>
</head>
<body>
	<div class="ashby-job-posting-container">
		<h1 class="ashby-job-posting-heading">Staff Backend Engineer</h1>
		<div class="ashby-job-posting-company">Acme Corp</div>
		<div class="ashby-job-posting-brief">Remote, US</div>
		<p>Compensation: $210,000 - $260,000 USD</p>
		<ul>
			<li>Lead backend architecture</li>
			<li>Mentor senior engineers</li>
		</ul>
	</div>
</body>
</html>`

		parsed := scraper.ParseJobPageHTML("https://jobs.ashbyhq.com/acme-corp/123", rawHTML)
		if parsed.RoleTitle != "Staff Backend Engineer" {
			t.Errorf("RoleTitle = %q, want %q", parsed.RoleTitle, "Staff Backend Engineer")
		}
		if parsed.CompanyName != "Acme Corp" {
			t.Errorf("CompanyName = %q, want %q", parsed.CompanyName, "Acme Corp")
		}
		if parsed.Location != "Remote, US" {
			t.Errorf("Location = %q, want %q", parsed.Location, "Remote, US")
		}
		if parsed.DetectedSource != "ashby" {
			t.Errorf("DetectedSource = %q, want %q", parsed.DetectedSource, "ashby")
		}
		if parsed.StatedMinComp != 210000 || parsed.StatedMaxComp != 260000 {
			t.Errorf("Comp = %d-%d, want 210000-260000", parsed.StatedMinComp, parsed.StatedMaxComp)
		}
	})

	t.Run("extracts title, company, and location from meta tags when JSON-LD is absent", func(t *testing.T) {
		rawHTML := `<!DOCTYPE html>
<html>
<head>
	<meta property="og:title" content="Principal Distributed Systems Engineer" />
	<meta property="og:site_name" content="Monzo" />
	<meta name="ashby:location" content="London, UK" />
</head>
<body>
	<div>
		<p>We are looking for a Principal Engineer in London.</p>
	</div>
</body>
</html>`

		parsed := scraper.ParseJobPageHTML("https://jobs.ashbyhq.com/monzo/abc-456", rawHTML)
		if parsed.RoleTitle != "Principal Distributed Systems Engineer" {
			t.Errorf("RoleTitle = %q, want %q", parsed.RoleTitle, "Principal Distributed Systems Engineer")
		}
		if parsed.CompanyName != "Monzo" {
			t.Errorf("CompanyName = %q, want %q", parsed.CompanyName, "Monzo")
		}
		if parsed.Location != "London, UK" {
			t.Errorf("Location = %q, want %q", parsed.Location, "London, UK")
		}
		if parsed.DetectedSource != "ashby" {
			t.Errorf("DetectedSource = %q, want %q", parsed.DetectedSource, "ashby")
		}
	})

	t.Run("extracts company from URL slug when site name tag is absent", func(t *testing.T) {
		rawHTML := `<!DOCTYPE html>
<html>
<head>
	<title>Infrastructure Lead</title>
	<meta name="location" content="San Francisco, CA" />
</head>
<body>
	<h1 class="_headingTitle_">Infrastructure Lead</h1>
</body>
</html>`

		parsed := scraper.ParseJobPageHTML("https://jobs.ashbyhq.com/stripe-payments/xyz", rawHTML)
		if parsed.RoleTitle != "Infrastructure Lead" {
			t.Errorf("RoleTitle = %q, want %q", parsed.RoleTitle, "Infrastructure Lead")
		}
		if parsed.CompanyName != "Stripe Payments" {
			t.Errorf("CompanyName = %q, want %q", parsed.CompanyName, "Stripe Payments")
		}
		if parsed.Location != "San Francisco, CA" {
			t.Errorf("Location = %q, want %q", parsed.Location, "San Francisco, CA")
		}
	})
}
