import { Component, signal } from '@angular/core';

@Component({
  selector: 'app-login-page',
  standalone: true,
  imports: [],
  templateUrl: './login-page.component.html',
  styleUrl: './login-page.component.scss',
})
export class LoginPageComponent {
  username = signal('');
  password = signal('');
  showPassword = signal(false);

  onSubmit(): void {
    // Design apenas - sem lógica de autenticação (será implementada em features futuras)
  }

  toggleShowPassword(): void {
    this.showPassword.update((v) => !v);
  }
}
