import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService, AuthValidator } from '../../services/auth.service';
import { ShieldStateService } from '../../services/shield-state.service';

@Component({
  selector: 'app-auth-modal',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="fixed inset-0 z-50 bg-slate-950/85 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="w-full max-w-[420px] rounded-2xl bg-slate-900 border border-slate-700 p-6 shadow-2xl space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-lg font-black text-white">🛡️ CoExist Shield AI</h2>
            <p class="text-xs font-bold text-sky-400 mt-0.5">
              {{ isRegister() ? 'Create New Account' : 'Sign In to Firebase Auth' }}
            </p>
          </div>
          <button
            type="button"
            (click)="state.showAuthModal.set(false)"
            class="text-slate-400 hover:text-white px-2 py-1 text-sm font-bold">
            ✕
          </button>
        </div>

        <div class="space-y-3">
          <div>
            <label class="block text-[11px] font-bold text-slate-300 mb-1">Email Address</label>
            <input
              type="email"
              [(ngModel)]="email"
              placeholder="ranger@coexistshield.ai"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border text-xs text-white focus:outline-none focus:border-sky-500"
              [ngClass]="emailError() ? 'border-red-500' : 'border-slate-700'" />
            @if (emailError(); as err) {
              <p class="text-[11px] text-red-400 mt-1">{{ err }}</p>
            }
          </div>

          <div>
            <label class="block text-[11px] font-bold text-slate-300 mb-1">Password (8+ chars)</label>
            <input
              type="password"
              [(ngModel)]="password"
              placeholder="••••••••"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border text-xs text-white focus:outline-none focus:border-sky-500"
              [ngClass]="passwordError() ? 'border-red-500' : 'border-slate-700'" />
            @if (passwordError(); as err) {
              <p class="text-[11px] text-red-400 mt-1">{{ err }}</p>
            }
          </div>

          @if (isRegister()) {
            <div>
              <label class="block text-[11px] font-bold text-slate-300 mb-1">Confirm Password</label>
              <input
                type="password"
                [(ngModel)]="confirm"
                placeholder="••••••••"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border text-xs text-white focus:outline-none focus:border-sky-500"
                [ngClass]="confirmError() ? 'border-red-500' : 'border-slate-700'" />
              @if (confirmError(); as err) {
                <p class="text-[11px] text-red-400 mt-1">{{ err }}</p>
              }
            </div>
          }
        </div>

        @if (serverError(); as sErr) {
          <div class="p-2.5 rounded-xl bg-red-950/70 border border-red-800 text-xs text-red-300 font-medium">
            {{ sErr }}
          </div>
        }

        @if (infoMessage(); as info) {
          <div class="p-2.5 rounded-xl bg-sky-950/70 border border-sky-800 text-xs text-sky-300 font-medium">
            {{ info }}
          </div>
        }

        <div class="space-y-2 pt-1">
          <button
            type="button"
            [disabled]="isLoading()"
            (click)="submit()"
            class="w-full py-2.5 px-4 rounded-xl bg-sky-600 hover:bg-sky-500 disabled:bg-slate-800 text-xs font-extrabold text-white transition-colors">
            {{ isLoading() ? 'Processing...' : isRegister() ? 'Register Account' : 'Sign In' }}
          </button>

          <div class="flex items-center justify-between pt-1">
            @if (!isRegister()) {
              <button
                type="button"
                [disabled]="isLoading()"
                (click)="resetPassword()"
                class="text-[11px] font-semibold text-slate-400 hover:text-sky-400">
                Forgot Password?
              </button>
            } @else {
              <span></span>
            }

            <button
              type="button"
              [disabled]="isLoading()"
              (click)="toggleMode()"
              class="text-[11px] font-semibold text-sky-400 hover:text-sky-300">
              {{ isRegister() ? 'Already have an account? Sign In' : 'New to CoExist? Register' }}
            </button>
          </div>

          <button
            type="button"
            [disabled]="isLoading()"
            (click)="state.showAuthModal.set(false)"
            class="w-full py-2 px-3 rounded-xl bg-slate-800 hover:bg-slate-700 text-[11px] font-bold text-slate-300">
            Continue as Guest (Demo Mode)
          </button>
        </div>
      </div>
    </div>
  `
})
export class AuthModalComponent {
  auth = inject(AuthService);
  state = inject(ShieldStateService);

  isRegister = signal<boolean>(false);
  isLoading = signal<boolean>(false);
  submitted = signal<boolean>(false);
  serverError = signal<string | null>(null);
  infoMessage = signal<string | null>(null);

  email = '';
  password = '';
  confirm = '';

  emailError(): string | null {
    return this.submitted() ? AuthValidator.validateEmail(this.email) : null;
  }

  passwordError(): string | null {
    return this.submitted() ? AuthValidator.validatePassword(this.password) : null;
  }

  confirmError(): string | null {
    return this.submitted() && this.isRegister()
      ? AuthValidator.validateConfirm(this.password, this.confirm)
      : null;
  }

  toggleMode() {
    this.isRegister.set(!this.isRegister());
    this.submitted.set(false);
    this.serverError.set(null);
    this.infoMessage.set(null);
  }

  async resetPassword() {
    this.submitted.set(true);
    this.serverError.set(null);
    this.infoMessage.set(null);
    if (AuthValidator.validateEmail(this.email)) return;

    this.isLoading.set(true);
    try {
      await this.auth.sendPasswordReset(this.email);
      this.infoMessage.set('Password reset email sent (if an account exists)');
    } catch (e: any) {
      this.serverError.set(e?.message || 'Failed to send password reset email');
    } finally {
      this.isLoading.set(false);
    }
  }

  async submit() {
    this.submitted.set(true);
    this.serverError.set(null);
    this.infoMessage.set(null);
    if (
      AuthValidator.validateEmail(this.email) ||
      AuthValidator.validatePassword(this.password) ||
      (this.isRegister() && AuthValidator.validateConfirm(this.password, this.confirm))
    ) {
      return;
    }

    this.isLoading.set(true);
    try {
      const user = this.isRegister()
        ? await this.auth.signUp(this.email, this.password)
        : await this.auth.signIn(this.email, this.password);
      this.state.onUserAuthenticated(user);
    } catch (e: any) {
      this.serverError.set(e?.message || 'Authentication failed');
    } finally {
      this.isLoading.set(false);
    }
  }
}
