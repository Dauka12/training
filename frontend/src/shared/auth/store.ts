import { create } from 'zustand';

export type AuthRole = 'user' | 'trainer' | 'admin';

export type AuthState = {
  isAuthenticated: boolean;
  role: AuthRole;
  email: string;
  onboardingDone: boolean;
  mustChangePassword: boolean;
  setAuthenticated: (payload: { role: AuthRole; email: string; onboardingDone: boolean; mustChangePassword?: boolean }) => void;
  clear: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  role: 'user',
  email: '',
  onboardingDone: false,
  mustChangePassword: false,
  setAuthenticated: ({ role, email, onboardingDone, mustChangePassword = false }) =>
    set({
      isAuthenticated: true,
      role,
      email,
      onboardingDone,
      mustChangePassword
    }),
  clear: () =>
    set({
      isAuthenticated: false,
      role: 'user',
      email: '',
      onboardingDone: false,
      mustChangePassword: false
    })
}));
