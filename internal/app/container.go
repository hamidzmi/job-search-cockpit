package app

import (
	"net/http"

	aiadapter "job-search-cockpit/internal/adapter/ai"
	clockadapter "job-search-cockpit/internal/adapter/clock"
	fsadapter "job-search-cockpit/internal/adapter/fs"
	httpadapter "job-search-cockpit/internal/adapter/http"
	"job-search-cockpit/internal/adapter/scraper"
	"job-search-cockpit/internal/adapter/sqlite"
	"job-search-cockpit/internal/usecase"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type Dependencies struct {
	Clock             outbound.Clock
	AssistantGateway  outbound.AssistantGateway
	JobPageParser     outbound.JobPageParser
	ArtifactBlobStore outbound.ArtifactBlobStore
}

func NewHTTPHandler(store *sqlite.Store, deps Dependencies) http.Handler {
	clock := deps.Clock
	if clock == nil {
		clock = clockadapter.SystemClock{}
	}
	assistant := deps.AssistantGateway
	if assistant == nil {
		assistant = aiadapter.NewAssistantGateway(nil)
	}
	parser := deps.JobPageParser
	if parser == nil {
		parser = scraper.NewJobPageParser(nil)
	}
	blobStore := deps.ArtifactBlobStore
	if blobStore == nil {
		blobStore = fsadapter.NewLocalArtifactBlobStore("")
	}

	companySvc := usecase.NewCompanyService(store, store, clock)
	opportunitySvc := usecase.NewOpportunityService(store, store, store, store, clock)
	aiSettingsSvc := usecase.NewAISettingsService(store, store, assistant, clock)
	quickCaptureSvc := usecase.NewQuickCaptureService(store, store, store, store, store, parser, assistant, clock)
	profileSvc := usecase.NewCandidateProfileService(store, store, clock)
	storySvc := usecase.NewStoryService(store, store, clock)
	contactSvc := usecase.NewContactService(store, store, store, store, clock)
	interactionSvc := usecase.NewInteractionService(store, store, store, store, store, clock)
	actionItemSvc := usecase.NewActionItemService(store, store, store, store, clock)
	interviewRoundSvc := usecase.NewInterviewRoundService(store, store, store, store, store, clock)
	fitAssessmentSvc := usecase.NewFitAssessmentService(store, store, store, store, store, store, assistant, clock)
	artifactSvc := usecase.NewApplicationArtifactService(store, store, store, store, store, store, store, assistant, blobStore, clock)

	return httpadapter.NewRouter(
		httpadapter.NewCompanyHandler(companySvc),
		httpadapter.NewOpportunityHandler(opportunitySvc),
		httpadapter.NewAISettingsHandler(aiSettingsSvc),
		httpadapter.NewQuickCaptureHandler(quickCaptureSvc),
		httpadapter.NewCandidateProfileHandler(profileSvc),
		httpadapter.NewStoryHandler(storySvc),
		httpadapter.NewContactHandler(contactSvc),
		httpadapter.NewInteractionHandler(interactionSvc),
		httpadapter.NewActionItemHandler(actionItemSvc),
		httpadapter.NewInterviewRoundHandler(interviewRoundSvc),
		httpadapter.NewFitAssessmentHandler(fitAssessmentSvc),
		httpadapter.NewApplicationArtifactHandler(artifactSvc),
	)
}
