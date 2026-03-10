import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface HelloResponse {
  message: string;
}

@Injectable({ providedIn: 'root' })
export class ApiService {
  constructor(private http: HttpClient) {}

  getHello(): Observable<HelloResponse> {
    return this.http.get<HelloResponse>('/api/hello', { withCredentials: true });
  }
}
