import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { supabase } from '@/integrations/supabase/client';
import type { Session } from '@supabase/supabase-js';

// Types for authentication
interface User {
  id: string;
  email: string;
  full_name: string;
  created_at: string;
  updated_at: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  register: (email: string, password: string, fullName: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [session, setSession] = useState<Session | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const isAuthenticated = !!user;

  const checkAuth = async () => {
    try {
      setIsLoading(true);
      const { data: { session } } = await supabase.auth.getSession();
      
      if (session?.user) {
        setSession(session);
        const { data: profile } = await supabase.from('profiles').select('*').eq('id', session.user.id).single();
        
        if (profile) {
          setUser({
            id: profile.id,
            email: profile.email || session.user.email || '',
            full_name: profile.full_name || '',
            created_at: profile.created_at,
            updated_at: profile.updated_at
          });
        }
      } else {
        setUser(null);
        setSession(null);
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      setUser(null);
      setSession(null);
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (email: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const { data, error } = await supabase.auth.signInWithPassword({ email, password });
      if (error) return { success: false, error: error.message };

      if (data.session?.user) {
        setSession(data.session);
        const { data: profile } = await supabase.from('profiles').select('*').eq('id', data.user.id).single();
        
        if (profile) {
          setUser({
            id: profile.id,
            email: profile.email || data.user.email || '',
            full_name: profile.full_name || '',
            created_at: profile.created_at,
            updated_at: profile.updated_at
          });
        }
        return { success: true };
      }

      return { success: false, error: 'Login failed' };
    } catch (error) {
      return { success: false, error: 'Network error. Please try again.' };
    }
  };

  const register = async (email: string, password: string, fullName: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const { data, error } = await supabase.auth.signUp({
        email,
        password,
        options: {
          emailRedirectTo: `${window.location.origin}/`,
          data: { full_name: fullName }
        }
      });

      if (error) return { success: false, error: error.message };
      if (data.session?.user) {
        setSession(data.session);
        const { data: profile } = await supabase.from('profiles').select('*').eq('id', data.user.id).single();
        
        if (profile) {
          setUser({
            id: profile.id,
            email: profile.email || data.user.email || '',
            full_name: profile.full_name || '',
            created_at: profile.created_at,
            updated_at: profile.updated_at
          });
        }
        return { success: true };
      }

      return { success: false, error: 'Registration failed' };
    } catch (error) {
      return { success: false, error: 'Network error. Please try again.' };
    }
  };

  const logout = async () => {
    try {
      await supabase.auth.signOut();
      setUser(null);
      setSession(null);
    } catch (error) {
      console.error('Logout error:', error);
    }
  };

  useEffect(() => {
    const { data: { subscription } } = supabase.auth.onAuthStateChange((event, session) => {
      setSession(session);
      
      if (session?.user) {
        setTimeout(async () => {
          const { data: profile } = await supabase.from('profiles').select('*').eq('id', session.user.id).single();
          
          if (profile) {
            setUser({
              id: profile.id,
              email: profile.email || session.user.email || '',
              full_name: profile.full_name || '',
              created_at: profile.created_at,
              updated_at: profile.updated_at
            });
          }
        }, 0);
      } else {
        setUser(null);
      }
    });

    checkAuth();
    return () => subscription.unsubscribe();
  }, []);

  return (
    <AuthContext.Provider value={{ user, isAuthenticated, isLoading, login, register, logout, checkAuth }}>
      {children}
    </AuthContext.Provider>
  );
};
