import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { CockpitApiProvider } from './api/CockpitApiContext';
import { AppRoutes } from './App';
import './styles.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <CockpitApiProvider>
        <BrowserRouter>
          <AppRoutes />
        </BrowserRouter>
      </CockpitApiProvider>
    </QueryClientProvider>
  </React.StrictMode>,
);
