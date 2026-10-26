import { requestJSON } from './httpClient';
import type { FlightDeckData } from './types';

export interface FlightDeckApi {
  getFlightDeck(): Promise<FlightDeckData>;
}

export const defaultFlightDeckApi: FlightDeckApi = {
  getFlightDeck: () => requestJSON<FlightDeckData>('/api/flight-deck'),
};
