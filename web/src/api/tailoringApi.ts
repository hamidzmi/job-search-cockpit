import { ApiError, requestJSON } from './httpClient';
import type {
  ApplicationArtifact,
  ApplicationArtifactUpsertInput,
  ArtifactDraft,
  ArtifactDraftInput,
  FitAssessment,
  FitAssessmentUpsertInput,
} from './types';

export interface TailoringApi {
  getFitAssessment(opportunityId: string): Promise<FitAssessment>;
  saveFitAssessment(
    opportunityId: string,
    input: FitAssessmentUpsertInput,
  ): Promise<FitAssessment>;
  generateFitAssessment(
    opportunityId: string,
    modelOverride?: string,
  ): Promise<FitAssessment>;
  listArtifacts(opportunityId: string): Promise<ApplicationArtifact[]>;
  createArtifact(
    opportunityId: string,
    input: ApplicationArtifactUpsertInput,
  ): Promise<ApplicationArtifact>;
  updateArtifact(
    opportunityId: string,
    artifactId: string,
    input: ApplicationArtifactUpsertInput,
  ): Promise<ApplicationArtifact>;
  deleteArtifact(opportunityId: string, artifactId: string): Promise<void>;
  draftArtifact(
    opportunityId: string,
    input: ArtifactDraftInput,
  ): Promise<ArtifactDraft>;
  uploadArtifactPDF(
    opportunityId: string,
    artifactId: string,
    file: File,
  ): Promise<ApplicationArtifact>;
}

export const defaultTailoringApi: TailoringApi = {
  getFitAssessment(opportunityId: string) {
    return requestJSON<FitAssessment>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/fit-assessment`,
    );
  },
  saveFitAssessment(opportunityId: string, input: FitAssessmentUpsertInput) {
    return requestJSON<FitAssessment>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/fit-assessment`,
      {
        method: 'PUT',
        body: JSON.stringify(input),
      },
    );
  },
  generateFitAssessment(opportunityId: string, modelOverride?: string) {
    return requestJSON<FitAssessment>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/fit-assessment/generate`,
      {
        method: 'POST',
        body: JSON.stringify({ model_override: modelOverride ?? '' }),
      },
    );
  },
  listArtifacts(opportunityId: string) {
    return requestJSON<ApplicationArtifact[]>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/artifacts`,
    );
  },
  createArtifact(opportunityId: string, input: ApplicationArtifactUpsertInput) {
    return requestJSON<ApplicationArtifact>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/artifacts`,
      {
        method: 'POST',
        body: JSON.stringify(input),
      },
    );
  },
  updateArtifact(
    opportunityId: string,
    artifactId: string,
    input: ApplicationArtifactUpsertInput,
  ) {
    return requestJSON<ApplicationArtifact>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/artifacts/${encodeURIComponent(artifactId)}`,
      {
        method: 'PUT',
        body: JSON.stringify(input),
      },
    );
  },
  deleteArtifact(opportunityId: string, artifactId: string) {
    return requestJSON<void>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/artifacts/${encodeURIComponent(artifactId)}`,
      {
        method: 'DELETE',
      },
    );
  },
  draftArtifact(opportunityId: string, input: ArtifactDraftInput) {
    return requestJSON<ArtifactDraft>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/artifacts/draft`,
      {
        method: 'POST',
        body: JSON.stringify(input),
      },
    );
  },
  async uploadArtifactPDF(
    opportunityId: string,
    artifactId: string,
    file: File,
  ): Promise<ApplicationArtifact> {
    const formData = new FormData();
    formData.append('file', file);
    const response = await fetch(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/artifacts/${encodeURIComponent(artifactId)}/file`,
      {
        method: 'POST',
        body: formData,
      },
    );
    if (!response.ok) {
      let message = `Request failed (${response.status})`;
      let code = 'api_error';
      try {
        const payload = await response.json();
        if (payload?.error?.message) {
          message = payload.error.message;
          code = payload.error.code ?? code;
        }
      } catch {
        // ignore non-json error bodies
      }
      throw new ApiError(message, code, response.status);
    }
    return (await response.json()) as ApplicationArtifact;
  },
};
