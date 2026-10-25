# Local-First Single-User Architecture

The Job Search Cockpit runs locally on the candidate's machine as a single-user application backed by a local SQLite database and local filesystem storage, with a Go backend and React frontend, rather than a cloud-hosted multi-tenant SaaS.

## Context & Trade-offs

Job search data contains sensitive personal data (compensation floors, private interview debriefs, candid notes on companies and contacts, and resume files). Running locally eliminates authentication, multi-tenant isolation, cloud hosting costs, and third-party storage of private career notes, while giving the Go backend direct access to local files and PDFs. In exchange, the application is designed for desktop/laptop use rather than out-of-the-box mobile access.
