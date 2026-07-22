export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: unknown
  ) {
    super(message);
    this.name = 'APIError';
  }
}

export interface APIResponse<T> {
  data: T;
  status: number;
}

// User
export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  company_name?: string;
  phone?: string;
  avatar?: string;
  plan_id: string;
  plan: 'free' | 'pro' | 'business';
  role: 'owner' | 'admin' | 'agent';
  is_active: boolean;
  must_change_password: boolean;
  onboarding_status?: string;
  industry?: string;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

// Auth
export interface LoginRequest {
  email: string;
  password: string;
}

export interface SignupRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  company_name?: string;
}

export interface AuthResponse {
  token?: string;
  refresh_token?: string;
  user: User;
  trial_info?: Record<string, unknown>;
}

// Conversation
export interface Conversation {
  id: string;
  customer_name: string;
  customer_avatar?: string;
  channel: 'whatsapp' | 'instagram' | 'telegram' | 'discord' | 'web';
  status: 'active' | 'resolved' | 'escalated';
  is_ai_transferred: boolean;
  last_message?: string;
  unread: number;
  intent?: string;
  priority: 'low' | 'medium' | 'high';
  created_at: string;
  updated_at: string;
}

export interface MessageMetadata {
  confidence?: number;
  language?: string;
  [key: string]: any;
}

export interface Message {
  id: string;
  conversation_id: string;
  content: string;
  sender_type?: 'customer' | 'ai' | 'agent' | 'system';
  role?: string;
  source?: string;
  metadata?: MessageMetadata;
  confidence?: number;
  created_at: string;
}

// Training
export interface Category {
  id: string;
  name: string;
  description?: string;
  color: string;
  qa_count: number;
  created_at: string;
  updated_at: string;
}

export interface QAPair {
  id: string;
  category_id: string;
  question: string;
  answer: string;
  created_at: string;
}

export interface UnknownQuestion {
  id: string;
  user_id?: string;
  question: string;
  conversation_id?: string;
  channel?: string;
  status: 'pending' | 'trained' | 'ignored';
  suggested_answer?: string;
  category_id?: string;
  created_at: string;
}

export interface UnknownQuestionListResponse {
  questions: UnknownQuestion[];
  total: number;
  limit: number;
  offset: number;
}

// Analytics
export interface OverviewStats {
  conversations_today: number;
  resolved_today: number;
  active_conversations: number;
  escalated_count: number;
  ai_resolution_rate: number;
  avg_response_time: number;
  satisfaction: number;
  total_conversations: number;
}

export interface TrendData {
  date: string;
  conversations: number;
}

export interface IntentData {
  intent: string;
  count: number;
}

export interface PeakHourData {
  hour: string;
  volume: number;
}

export interface ChannelDistribution {
  [channel: string]: number;
}

// Integration
export interface Integration {
  id?: string;
  channel: string;
  status: string;
  config?: Record<string, unknown>;
  webhook_url?: string;
  connected_at?: string;
  created_at?: string;
  updated_at?: string;
  last_error?: string | null;
}

// Team
export interface TeamMember {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: 'owner' | 'admin' | 'agent';
  status: 'active' | 'pending';
}

// Plan
export interface Plan {
  id: string;
  name: string;
  price_ngn: number;
  features: string[];
  is_popular: boolean;
  max_responses: number;
  max_channels: number;
  max_team_members: number;
}

// API Key
export interface APIKey {
  id: string;
  name: string;
  key?: string;
  created_at: string;
}

// Handoff / Leads
export interface Handoff {
  id: string
  conversation_id: string
  customer_name: string
  customer_phone: string
  customer_whatsapp: string
  customer_location: string
  product_name: string
  original_price: number
  agreed_price: number
  quantity: number
  status: 'pending' | 'sold' | 'lost' | 'expired'
  final_price: number | null
  owner_notes: string
  reminder_count: number
  created_at: string
}

export interface HandoffsResponse {
  handoffs: Handoff[]
}

// Billing System
export interface UserCredit {
  id: string;
  user_id: string;
  balance: number;
  expires_at: string;
  last_updated_at: string;
}

export interface CreditPurchase {
  id: string;
  user_id: string;
  checkout_id: string;
  pack_type: string;
  amount: number;
  status: string;
  purchased_at: string;
  expires_at: string;
}

export interface CampaignSchedule {
  id: string;
  user_id: string;
  name: string;
  start_date: string;
  end_date: string;
  status: 'draft' | 'active' | 'completed' | 'cancelled';
  created_at: string;
  updated_at: string;
}

export interface PlanLimit {
  plan_id: string;
  max_responses: number;
  max_handoffs: number;
  max_inventory_items: number;
  has_notification: boolean;
  price_ngn: number;
  description: string;
}

// Create Campaign Request
export interface CreateCampaignRequest {
  name: string;
  start_date: string; // ISO date string
  end_date: string;   // ISO date string
}

// WebSocket
export interface WSMessage {
  type: 'new_message' | 'new_conversation' | 'status_change' | 'typing' | 'integration_update' | 'unknown_question' | 'notification' | 'typing_indicator';
  conversation_id?: string;
  content?: string;
  sender_type?: string;
  timestamp?: string;
  data?: any;
}
