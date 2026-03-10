import { Component, inject, OnInit, signal } from '@angular/core';
import { ApiService } from './core/api.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [],
  templateUrl: './app.component.html',
  styles: [
    `
      .hello-world {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        font-family: system-ui, sans-serif;
      }
      h1 {
        font-size: 2.5rem;
        margin: 0;
      }
      p {
        color: #666;
        margin-top: 0.5rem;
      }
      .loading {
        color: #999;
      }
      .error {
        color: #c00;
      }
      a {
        color: #06c;
        margin-top: 1rem;
      }
    `,
  ],
})
export class AppComponent implements OnInit {
  private api = inject(ApiService);

  message = signal<string>('');
  loading = signal(true);
  error = signal<string | null>(null);

  ngOnInit() {
    this.api.getHello().subscribe({
      next: (res) => {
        this.message.set(res.message);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
        this.error.set('Não autenticado. Redirecionando para login...');
        window.location.href = '/login';
      },
    });
  }
}
