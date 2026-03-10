import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, switchMap, throwError } from 'rxjs';
import { HttpClient } from '@angular/common/http';

const LOGIN_URL = '/login';
const REFRESH_URL = '/api/auth/refresh';

function isRefreshOrLoginUrl(url: string): boolean {
  const path = url.includes('http') ? new URL(url).pathname : url;
  return path === '/login' || path.endsWith('/api/auth/refresh');
}

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const http = inject(HttpClient);

  return next(req).pipe(
    catchError((err: HttpErrorResponse) => {
      if (err.status !== 401) {
        return throwError(() => err);
      }
      if (isRefreshOrLoginUrl(req.url)) {
        return throwError(() => err);
      }

      return http
        .post(REFRESH_URL, {}, { withCredentials: true })
        .pipe(
          switchMap(() => next(req)),
          catchError((refreshErr) => {
            if (
              refreshErr?.status === 401 ||
              refreshErr?.status === 402
            ) {
              window.location.href = LOGIN_URL;
            }
            return throwError(() => refreshErr ?? err);
          })
        );
    })
  );
};
