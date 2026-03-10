import {
  ApplicationConfig,
  provideZoneChangeDetection,
} from '@angular/core';
import {
  provideHttpClient,
  withInterceptors,
  withFetch,
} from '@angular/common/http';
import { authInterceptor } from './core/auth.interceptor';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideHttpClient(
      withFetch(),
      withInterceptors([authInterceptor])
    ),
  ],
};
