import { InjectionToken } from '@angular/core';

/**
 * Base URL for API requests. Empty string for relative URLs (proxy in dev, same-origin in prod).
 * Override via provider when deploying to different domains.
 */
export const API_BASE_URL = new InjectionToken<string>('API_BASE_URL', {
  providedIn: 'root',
  factory: () => '',
});
