import { create } from 'zustand';

export type AuthRole = 'user' | 'trainer' | 'admin';

export type AuthState = {
  isAuthenticated: boolean;
  role: AuthRole;
  email: string;
  onboardingDone: boolean;
  setAuthenticated: (payload: { role: AuthRole; email: string; onboardingDone: boolean }) => void;
  clear: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  role: 'user',
  email: '',
  onboardingDone: false,
  setAuthenticated: ({ role, email, onboardingDone }) =>
    set({
      isAuthenticated: true,
      role,
      email,
      onboardingDone
    }),
  clear: () =>
    set({
      isAuthenticated: false,
      role: 'user',
      email: '',
      onboardingDone: false
    })
}));
