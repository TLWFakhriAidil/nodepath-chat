export type Json =
  | string
  | number
  | boolean
  | null
  | { [key: string]: Json | undefined }
  | Json[]

export type Database = {
  // Allows to automatically instantiate createClient with right options
  // instead of createClient<Database, { PostgrestVersion: 'XX' }>(URL, KEY)
  __InternalSupabase: {
    PostgrestVersion: "13.0.5"
  }
  public: {
    Tables: {
      ai_conversations: {
        Row: {
          conversation_current: string | null
          conversation_last: string | null
          created_at: string | null
          current_node_id: string | null
          device_id: string | null
          flow_id: string | null
          human_status: Database["public"]["Enums"]["human_status"] | null
          id: string
          prospect_name: string | null
          prospect_num: string
          session_locked: boolean | null
          session_locked_at: string | null
          stage: string | null
          updated_at: string | null
          user_id: string
          waiting_for_reply: boolean | null
        }
        Insert: {
          conversation_current?: string | null
          conversation_last?: string | null
          created_at?: string | null
          current_node_id?: string | null
          device_id?: string | null
          flow_id?: string | null
          human_status?: Database["public"]["Enums"]["human_status"] | null
          id?: string
          prospect_name?: string | null
          prospect_num: string
          session_locked?: boolean | null
          session_locked_at?: string | null
          stage?: string | null
          updated_at?: string | null
          user_id: string
          waiting_for_reply?: boolean | null
        }
        Update: {
          conversation_current?: string | null
          conversation_last?: string | null
          created_at?: string | null
          current_node_id?: string | null
          device_id?: string | null
          flow_id?: string | null
          human_status?: Database["public"]["Enums"]["human_status"] | null
          id?: string
          prospect_name?: string | null
          prospect_num?: string
          session_locked?: boolean | null
          session_locked_at?: string | null
          stage?: string | null
          updated_at?: string | null
          user_id?: string
          waiting_for_reply?: boolean | null
        }
        Relationships: [
          {
            foreignKeyName: "ai_conversations_device_id_fkey"
            columns: ["device_id"]
            isOneToOne: false
            referencedRelation: "device_settings"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "ai_conversations_flow_id_fkey"
            columns: ["flow_id"]
            isOneToOne: false
            referencedRelation: "chatbot_flows"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "ai_conversations_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      analytics_summary: {
        Row: {
          active_users: number | null
          avg_response_time: unknown | null
          created_at: string | null
          date_period: string
          device_id: string | null
          id: string
          total_conversations: number | null
          total_messages: number | null
          updated_at: string | null
          user_id: string
        }
        Insert: {
          active_users?: number | null
          avg_response_time?: unknown | null
          created_at?: string | null
          date_period: string
          device_id?: string | null
          id?: string
          total_conversations?: number | null
          total_messages?: number | null
          updated_at?: string | null
          user_id: string
        }
        Update: {
          active_users?: number | null
          avg_response_time?: unknown | null
          created_at?: string | null
          date_period?: string
          device_id?: string | null
          id?: string
          total_conversations?: number | null
          total_messages?: number | null
          updated_at?: string | null
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "analytics_summary_device_id_fkey"
            columns: ["device_id"]
            isOneToOne: false
            referencedRelation: "device_settings"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "analytics_summary_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      chatbot_flows: {
        Row: {
          created_at: string | null
          description: string | null
          device_id: string | null
          edges: Json | null
          flow_id: string
          flow_name: string
          global_instance: string | null
          global_open_router_key: string | null
          id: string
          niche: string | null
          nodes: Json | null
          status: Database["public"]["Enums"]["flow_status"] | null
          updated_at: string | null
          user_id: string
        }
        Insert: {
          created_at?: string | null
          description?: string | null
          device_id?: string | null
          edges?: Json | null
          flow_id: string
          flow_name: string
          global_instance?: string | null
          global_open_router_key?: string | null
          id?: string
          niche?: string | null
          nodes?: Json | null
          status?: Database["public"]["Enums"]["flow_status"] | null
          updated_at?: string | null
          user_id: string
        }
        Update: {
          created_at?: string | null
          description?: string | null
          device_id?: string | null
          edges?: Json | null
          flow_id?: string
          flow_name?: string
          global_instance?: string | null
          global_open_router_key?: string | null
          id?: string
          niche?: string | null
          nodes?: Json | null
          status?: Database["public"]["Enums"]["flow_status"] | null
          updated_at?: string | null
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "chatbot_flows_device_id_fkey"
            columns: ["device_id"]
            isOneToOne: false
            referencedRelation: "device_settings"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "chatbot_flows_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      conversation_logs: {
        Row: {
          ai_response_time: unknown | null
          content: string | null
          created_at: string | null
          device_id: string | null
          id: string
          message_type: Database["public"]["Enums"]["message_type"] | null
          phone_number: string
          user_id: string
        }
        Insert: {
          ai_response_time?: unknown | null
          content?: string | null
          created_at?: string | null
          device_id?: string | null
          id?: string
          message_type?: Database["public"]["Enums"]["message_type"] | null
          phone_number: string
          user_id: string
        }
        Update: {
          ai_response_time?: unknown | null
          content?: string | null
          created_at?: string | null
          device_id?: string | null
          id?: string
          message_type?: Database["public"]["Enums"]["message_type"] | null
          phone_number?: string
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "conversation_logs_device_id_fkey"
            columns: ["device_id"]
            isOneToOne: false
            referencedRelation: "device_settings"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "conversation_logs_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      device_settings: {
        Row: {
          api_key_option: string | null
          api_provider: Database["public"]["Enums"]["api_provider"] | null
          created_at: string | null
          device_id: string
          device_name: string
          id: string
          status: Database["public"]["Enums"]["device_status"] | null
          updated_at: string | null
          user_id: string
          waha_url: string | null
          whacenter_token: string | null
          whatsapp_provider:
            | Database["public"]["Enums"]["whatsapp_provider"]
            | null
        }
        Insert: {
          api_key_option?: string | null
          api_provider?: Database["public"]["Enums"]["api_provider"] | null
          created_at?: string | null
          device_id: string
          device_name: string
          id?: string
          status?: Database["public"]["Enums"]["device_status"] | null
          updated_at?: string | null
          user_id: string
          waha_url?: string | null
          whacenter_token?: string | null
          whatsapp_provider?:
            | Database["public"]["Enums"]["whatsapp_provider"]
            | null
        }
        Update: {
          api_key_option?: string | null
          api_provider?: Database["public"]["Enums"]["api_provider"] | null
          created_at?: string | null
          device_id?: string
          device_name?: string
          id?: string
          status?: Database["public"]["Enums"]["device_status"] | null
          updated_at?: string | null
          user_id?: string
          waha_url?: string | null
          whacenter_token?: string | null
          whatsapp_provider?:
            | Database["public"]["Enums"]["whatsapp_provider"]
            | null
        }
        Relationships: [
          {
            foreignKeyName: "device_settings_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      profiles: {
        Row: {
          avatar_url: string | null
          created_at: string | null
          email: string | null
          full_name: string | null
          id: string
          updated_at: string | null
        }
        Insert: {
          avatar_url?: string | null
          created_at?: string | null
          email?: string | null
          full_name?: string | null
          id: string
          updated_at?: string | null
        }
        Update: {
          avatar_url?: string | null
          created_at?: string | null
          email?: string | null
          full_name?: string | null
          id?: string
          updated_at?: string | null
        }
        Relationships: []
      }
    }
    Views: {
      [_ in never]: never
    }
    Functions: {
      [_ in never]: never
    }
    Enums: {
      api_provider: "openai" | "anthropic" | "google" | "openrouter"
      device_status: "active" | "inactive" | "pending"
      flow_status: "active" | "inactive" | "draft"
      human_status: "active" | "inactive" | "pending_response"
      message_type: "text" | "image" | "audio" | "video" | "document"
      whatsapp_provider: "waha" | "whacenter" | "wablas" | "wasapbot"
    }
    CompositeTypes: {
      [_ in never]: never
    }
  }
}

type DatabaseWithoutInternals = Omit<Database, "__InternalSupabase">

type DefaultSchema = DatabaseWithoutInternals[Extract<keyof Database, "public">]

export type Tables<
  DefaultSchemaTableNameOrOptions extends
    | keyof (DefaultSchema["Tables"] & DefaultSchema["Views"])
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof (DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"] &
        DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Views"])
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? (DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"] &
      DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Views"])[TableName] extends {
      Row: infer R
    }
    ? R
    : never
  : DefaultSchemaTableNameOrOptions extends keyof (DefaultSchema["Tables"] &
        DefaultSchema["Views"])
    ? (DefaultSchema["Tables"] &
        DefaultSchema["Views"])[DefaultSchemaTableNameOrOptions] extends {
        Row: infer R
      }
      ? R
      : never
    : never

export type TablesInsert<
  DefaultSchemaTableNameOrOptions extends
    | keyof DefaultSchema["Tables"]
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"]
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"][TableName] extends {
      Insert: infer I
    }
    ? I
    : never
  : DefaultSchemaTableNameOrOptions extends keyof DefaultSchema["Tables"]
    ? DefaultSchema["Tables"][DefaultSchemaTableNameOrOptions] extends {
        Insert: infer I
      }
      ? I
      : never
    : never

export type TablesUpdate<
  DefaultSchemaTableNameOrOptions extends
    | keyof DefaultSchema["Tables"]
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"]
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"][TableName] extends {
      Update: infer U
    }
    ? U
    : never
  : DefaultSchemaTableNameOrOptions extends keyof DefaultSchema["Tables"]
    ? DefaultSchema["Tables"][DefaultSchemaTableNameOrOptions] extends {
        Update: infer U
      }
      ? U
      : never
    : never

export type Enums<
  DefaultSchemaEnumNameOrOptions extends
    | keyof DefaultSchema["Enums"]
    | { schema: keyof DatabaseWithoutInternals },
  EnumName extends DefaultSchemaEnumNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaEnumNameOrOptions["schema"]]["Enums"]
    : never = never,
> = DefaultSchemaEnumNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaEnumNameOrOptions["schema"]]["Enums"][EnumName]
  : DefaultSchemaEnumNameOrOptions extends keyof DefaultSchema["Enums"]
    ? DefaultSchema["Enums"][DefaultSchemaEnumNameOrOptions]
    : never

export type CompositeTypes<
  PublicCompositeTypeNameOrOptions extends
    | keyof DefaultSchema["CompositeTypes"]
    | { schema: keyof DatabaseWithoutInternals },
  CompositeTypeName extends PublicCompositeTypeNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[PublicCompositeTypeNameOrOptions["schema"]]["CompositeTypes"]
    : never = never,
> = PublicCompositeTypeNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[PublicCompositeTypeNameOrOptions["schema"]]["CompositeTypes"][CompositeTypeName]
  : PublicCompositeTypeNameOrOptions extends keyof DefaultSchema["CompositeTypes"]
    ? DefaultSchema["CompositeTypes"][PublicCompositeTypeNameOrOptions]
    : never

export const Constants = {
  public: {
    Enums: {
      api_provider: ["openai", "anthropic", "google", "openrouter"],
      device_status: ["active", "inactive", "pending"],
      flow_status: ["active", "inactive", "draft"],
      human_status: ["active", "inactive", "pending_response"],
      message_type: ["text", "image", "audio", "video", "document"],
      whatsapp_provider: ["waha", "whacenter", "wablas", "wasapbot"],
    },
  },
} as const
