import React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { AppShell } from './components/layout/AppShell';
import { AnalyticsWorkspace } from './features/analytics/AnalyticsWorkspace';
import { FlightDeckWorkspace } from './features/flightdeck/FlightDeckWorkspace';
import { NetworkWorkspace } from './features/network/NetworkWorkspace';
import { OpportunityDetailWorkspace } from './features/pipeline/OpportunityDetailWorkspace';
import { PipelineWorkspace } from './features/pipeline/PipelineWorkspace';
import { ProfileWorkspace } from './features/profile/ProfileWorkspace';
import { SettingsWorkspace } from './features/settings/SettingsWorkspace';

export function AppRoutes() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route path="/" element={<FlightDeckWorkspace />} />
        <Route path="/opportunities" element={<PipelineWorkspace />} />
        <Route path="/opportunities/:id" element={<OpportunityDetailWorkspace />} />
        <Route path="/network" element={<NetworkWorkspace />} />
        <Route path="/profile" element={<ProfileWorkspace />} />
        <Route path="/analytics" element={<AnalyticsWorkspace />} />
        <Route path="/settings" element={<SettingsWorkspace />} />
        <Route path="*" element={<Navigate to="/opportunities" replace />} />
      </Route>
    </Routes>
  );
}
