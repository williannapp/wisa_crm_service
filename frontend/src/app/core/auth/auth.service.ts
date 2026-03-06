import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { API_BASE_URL } from '../http/api-config';

export interface LoginRequest {
  tenant_slug: string;
  product_slug: string;
  user_email: string;
  password: string;
  state: string;
}

export interface LoginResponse {
  redirect_url: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly apiBase = inject(API_BASE_URL);

  login(credentials: LoginRequest): Observable<LoginResponse> {
    const url = `${this.apiBase}/api/v1/auth/login`;
    return this.http.post<LoginResponse>(url, credentials, {
      withCredentials: true,
    });
  }
}
