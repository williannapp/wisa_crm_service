import {
  Component,
  computed,
  DestroyRef,
  inject,
  OnInit,
  signal,
} from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';

import { AuthService, LoginRequest } from '../../../core/auth/auth.service';

@Component({
  selector: 'app-login-page',
  standalone: true,
  imports: [],
  templateUrl: './login-page.component.html',
  styleUrl: './login-page.component.scss',
})
export class LoginPageComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly authService = inject(AuthService);
  private readonly destroyRef = inject(DestroyRef);

  tenantSlug = signal('');
  productSlug = signal('');
  state = signal('');
  queryParamsError = signal<string | null>(null);
  isParamsValid = computed(() => this.queryParamsError() === null);

  userEmail = signal('');
  password = signal('');
  showPassword = signal(false);
  isSubmitting = signal(false);
  loginError = signal<string | null>(null);

  ngOnInit(): void {
    const params = this.route.snapshot.queryParams;
    const tenant = (params['tenant_slug'] ?? '').trim();
    const product = (params['product_slug'] ?? '').trim();
    const stateVal = (params['state'] ?? '').trim();

    this.tenantSlug.set(tenant);
    this.productSlug.set(product);
    this.state.set(stateVal);

    if (!tenant || !product || !stateVal) {
      this.queryParamsError.set(
        'Parâmetros inválidos. Acesso deve ser feito através do sistema do cliente.'
      );
    }
  }

  onSubmit(): void {
    if (this.queryParamsError()) return;
    if (this.isSubmitting()) return;

    const body: LoginRequest = {
      tenant_slug: this.tenantSlug(),
      product_slug: this.productSlug(),
      user_email: this.userEmail(),
      password: this.password(),
      state: this.state(),
    };

    this.loginError.set(null);
    this.isSubmitting.set(true);

    this.authService
      .login(body)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (res) => {
          window.location.href = res.redirect_url;
        },
        error: () => {
          this.isSubmitting.set(false);
          this.loginError.set('Credenciais inválidas. Tente novamente.');
        },
      });
  }

  toggleShowPassword(): void {
    this.showPassword.update((v) => !v);
  }
}
