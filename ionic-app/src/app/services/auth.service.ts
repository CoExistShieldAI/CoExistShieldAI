import { Injectable, signal } from '@angular/core';
import { AuthUser, UserRole } from '../models/shield.model';

const REFRESH_TOKEN_KEY = 'coexistshield.refresh_token';
const BASE_URL = 'https://identitytoolkit.googleapis.com/v1';
const SECURE_TOKEN_URL = 'https://securetoken.googleapis.com/v1/token';

export class AuthValidator {
  static readonly MIN_PASSWORD_LENGTH = 8;
  private static readonly EMAIL_REGEX = /^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$/;

  static validateEmail(email: string): string | null {
    if (!email || !email.trim()) return 'Please enter your email address';
    if (!this.EMAIL_REGEX.test(email.trim())) return 'Invalid email address format';
    return null;
  }

  static validatePassword(password: string): string | null {
    if (!password) return 'Please enter a password';
    if (password.length < this.MIN_PASSWORD_LENGTH) {
      return `Password must be at least ${this.MIN_PASSWORD_LENGTH} characters`;
    }
    return null;
  }

  static validateConfirm(password: string, confirm: string): string | null {
    return password !== confirm ? 'Passwords do not match' : null;
  }
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  readonly currentUser = signal<AuthUser | null>(null);

  private getApiKey(): string {
    if (typeof window === 'undefined') return '';
    const cfg = (window as any).__COEXIST_CONFIG__;
    return cfg?.firebaseApiKey || cfg?.googleMapsApiKey || '';
  }

  async restoreSession(): Promise<AuthUser | null> {
    if (typeof window === 'undefined') return null;
    const savedRefresh = localStorage.getItem(REFRESH_TOKEN_KEY);
    if (!savedRefresh) return null;
    try {
      const user = await this.refresh(savedRefresh);
      this.currentUser.set(user);
      localStorage.setItem(REFRESH_TOKEN_KEY, user.refreshToken);
      return user;
    } catch {
      localStorage.removeItem(REFRESH_TOKEN_KEY);
      return null;
    }
  }

  async signIn(email: string, password: string): Promise<AuthUser> {
    return this.authenticate('accounts:signInWithPassword', email, password);
  }

  async signUp(email: string, password: string): Promise<AuthUser> {
    return this.authenticate('accounts:signUp', email, password);
  }

  async sendPasswordReset(email: string): Promise<void> {
    const apiKey = this.getApiKey();
    if (!apiKey) {
      throw new Error('Authentication is not configured (Firebase API key missing)');
    }
    const res = await fetch(`${BASE_URL}/accounts:sendOobCode?key=${apiKey}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ requestType: 'PASSWORD_RESET', email: email.trim() })
    });
    if (!res.ok) {
      const errBody = await res.json().catch(() => ({}));
      const code: string = errBody?.error?.message || '';
      if (code.startsWith('EMAIL_NOT_FOUND')) {
        return; // Prevent account enumeration
      }
      throw new Error(this.translateError(code));
    }
  }

  signOut(): void {
    this.currentUser.set(null);
    if (typeof window !== 'undefined') {
      localStorage.removeItem(REFRESH_TOKEN_KEY);
    }
  }

  private async authenticate(endpoint: string, email: string, password: string): Promise<AuthUser> {
    const apiKey = this.getApiKey();
    if (!apiKey) {
      throw new Error('Authentication is not configured (Firebase API key missing)');
    }
    const res = await fetch(`${BASE_URL}/${endpoint}?key=${apiKey}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: email.trim(),
        password,
        returnSecureToken: true
      })
    });
    if (!res.ok) {
      const errBody = await res.json().catch(() => ({}));
      const code: string = errBody?.error?.message || '';
      throw new Error(this.translateError(code));
    }
    const body = await res.json();
    const user: AuthUser = {
      uid: body.localId,
      email: body.email || email.trim(),
      role: this.roleFromIdToken(body.idToken),
      idToken: body.idToken,
      refreshToken: body.refreshToken
    };
    this.currentUser.set(user);
    if (typeof window !== 'undefined') {
      localStorage.setItem(REFRESH_TOKEN_KEY, user.refreshToken);
    }
    return user;
  }

  private async refresh(refreshToken: string): Promise<AuthUser> {
    const apiKey = this.getApiKey();
    if (!apiKey) {
      throw new Error('Authentication is not configured');
    }
    const params = new URLSearchParams();
    params.set('grant_type', 'refresh_token');
    params.set('refresh_token', refreshToken);

    const res = await fetch(`${SECURE_TOKEN_URL}?key=${apiKey}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: params.toString()
    });
    if (!res.ok) {
      throw new Error('Session has expired. Please sign in again.');
    }
    const body = await res.json();
    return {
      uid: body.user_id,
      email: this.emailFromIdToken(body.id_token),
      role: this.roleFromIdToken(body.id_token),
      idToken: body.id_token,
      refreshToken: body.refresh_token
    };
  }

  private claimsOf(idToken: string): Record<string, any> | null {
    try {
      const payload = idToken.split('.')[1];
      if (!payload) return null;
      const base64 = payload.replace(/-/g, '+').replace(/_/g, '/');
      const json = atob(base64);
      return JSON.parse(json);
    } catch {
      return null;
    }
  }

  private emailFromIdToken(idToken: string): string {
    return this.claimsOf(idToken)?.['email'] || '';
  }

  private roleFromIdToken(idToken: string): UserRole {
    const claim = (this.claimsOf(idToken)?.['role'] || '').toUpperCase();
    if (claim === 'ADMIN' || claim === 'OFFICIAL' || claim === 'CITIZEN') {
      return claim as UserRole;
    }
    return 'CITIZEN';
  }

  private translateError(code: string): string {
    if (code.startsWith('EMAIL_EXISTS')) return 'This email address is already registered';
    if (code.startsWith('WEAK_PASSWORD')) return 'Password is too weak (must be at least 8 characters)';
    if (code.startsWith('INVALID_EMAIL')) return 'Invalid email address format';
    if (
      code.startsWith('EMAIL_NOT_FOUND') ||
      code.startsWith('INVALID_PASSWORD') ||
      code.startsWith('INVALID_LOGIN_CREDENTIALS')
    ) {
      return 'Invalid email address or password';
    }
    if (code.startsWith('USER_DISABLED')) return 'This account has been disabled';
    if (code.startsWith('TOO_MANY_ATTEMPTS_TRY_LATER')) return 'Too many attempts. Please try again later';
    if (code.startsWith('OPERATION_NOT_ALLOWED')) return 'Email/Password sign-in is not enabled';
    if (code.startsWith('CONFIGURATION_NOT_FOUND')) return 'Authentication provider configuration not found';
    return `Authentication failed${code ? ` (${code})` : ''}`;
  }
}
