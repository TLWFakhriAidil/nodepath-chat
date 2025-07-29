import { createClient, SupabaseClient } from '@supabase/supabase-js'

export type MediaFile = {
  id: string
  filename: string
  original_name: string
  file_type: string
  file_size: number
  storage_path: string
  public_url: string
  uploaded_at: string
}

// Helper function to check if Supabase is configured
export const isSupabaseConfigured = (): boolean => {
  try {
    const supabaseUrl = import.meta.env?.VITE_SUPABASE_URL
    const supabaseAnonKey = import.meta.env?.VITE_SUPABASE_ANON_KEY
    return !!(supabaseUrl && supabaseAnonKey)
  } catch (error) {
    return false
  }
}

// Function to get Supabase client (lazy initialization)
export const getSupabaseClient = (): SupabaseClient | null => {
  try {
    const supabaseUrl = import.meta.env?.VITE_SUPABASE_URL
    const supabaseAnonKey = import.meta.env?.VITE_SUPABASE_ANON_KEY
    
    if (!supabaseUrl || !supabaseAnonKey) {
      return null
    }
    
    return createClient(supabaseUrl, supabaseAnonKey)
  } catch (error) {
    console.warn('Failed to initialize Supabase client:', error)
    return null
  }
}

// Export a getter for the supabase client
export const supabase = getSupabaseClient()