export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  plan_id: string;
  status: string;
  created_at: string;
  last_login_at: string;
  avatar_url?: string;
  org_id?: string;
  org_name?: string;
  total_conversations: number;
  total_messages: number;
  credits_remaining: number;
  health_score: number;
}

export interface Conversation {
  id: string;
  user_id: string;
  customer_name: string;
  customer_phone: string;
  status: string;
  ai_handled: boolean;
  sentiment: 'positive' | 'neutral' | 'negative';
  message_count: number;
  created_at: string;
  last_message_at: string;
  last_message: string;
}

export interface RevenueData {
  mrr: number;
  arr: number;
  total_revenue: number;
  paying_users: number;
  churn_rate: number;
  ltv: number;
  mrr_history: MonthlyRevenue[];
  plan_breakdown: PlanRevenue[];
  failed_payments: FailedPayment[];
}

export interface MonthlyRevenue {
  month: string;
  revenue: number;
  users: number;
}

export interface PlanRevenue {
  plan: string;
  users: number;
  revenue: number;
  percentage: number;
}

export interface FailedPayment {
  id: string;
  user_id: string;
  user_email: string;
  amount: number;
  reason: string;
  created_at: string;
}

export interface AnalyticsData {
  visitors_today: number;
  visitors_yesterday: number;
  signups_today: number;
  conversion_rate: number;
  bounce_rate: number;
  avg_session_duration: number;
  page_views: PageView[];
  traffic_sources: TrafficSource[];
  visitor_history: DailyVisitors[];
  funnel: FunnelStep[];
}

export interface PageView {
  path: string;
  views: number;
  unique_visitors: number;
  avg_time_on_page: number;
  bounce_rate: number;
}

export interface TrafficSource {
  source: string;
  visitors: number;
  signups: number;
  conversion: number;
}

export interface DailyVisitors {
  date: string;
  visitors: number;
  signups: number;
}

export interface FunnelStep {
  step: string;
  count: number;
  percentage: number;
  dropoff: number;
}

export interface AIHealthData {
  accuracy: number;
  accuracy_trend: number;
  total_queries: number;
  answered_correctly: number;
  unanswered_questions: UnansweredQuestion[];
  accuracy_history: DailyAccuracy[];
  sentiment_breakdown: SentimentBreakdown;
}

export interface UnansweredQuestion {
  question: string;
  count: number;
  last_seen: string;
  suggested_answer?: string;
}

export interface DailyAccuracy {
  date: string;
  accuracy: number;
  queries: number;
}

export interface SentimentBreakdown {
  positive: number;
  neutral: number;
  negative: number;
}

export interface SystemHealth {
  api: ServiceStatus;
  database: ServiceStatus;
  redis: ServiceStatus;
  whatsapp: ServiceStatus;
  error_rate: number;
  p50_latency: number;
  p95_latency: number;
  p99_latency: number;
  active_websockets: number;
  job_queue_depth: number;
}

export interface ServiceStatus {
  name: string;
  status: 'healthy' | 'degraded' | 'down';
  latency_ms: number;
  last_check: string;
  uptime: number;
}

export interface Alert {
  id: string;
  type: 'critical' | 'warning' | 'info';
  title: string;
  message: string;
  timestamp: string;
  acknowledged: boolean;
}

export interface LiveFeedEvent {
  id: string;
  type: 'signup' | 'payment' | 'escalation' | 'ai_failure' | 'whatsapp_issue' | 'system';
  title: string;
  description: string;
  timestamp: string;
  severity: 'high' | 'medium' | 'low';
}

export interface AdminUser {
  id: string;
  email: string;
  role: 'owner' | 'admin' | 'support' | 'readonly';
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: AdminUser & { first_name: string; last_name: string };
  trial_info?: {
    trial_expires_at?: string;
    trial_ended?: boolean;
    trial_days_left?: number;
  };
}

export interface OverviewResponse {
  total_users: number;
  paying_users: number;
  active_users: number;
  total_revenue: number;
  mrr: number;
  churn_rate: number;
  total_conversations: number;
  system_status: string;
}

export interface UsersResponse {
  users: User[];
  total: number;
}

export interface UserDetail extends User {
  total_conversations: number;
  total_messages: number;
  credits_remaining: number;
  health_score: number;
}

export interface AnalyticsResponse {
  visitors_today: number;
  visitors_yesterday: number;
  signups_today: number;
  conversion_rate: number;
  total_signups: number;
  bounce_rate: number;
  avg_session_duration: number;
  page_views: PageView[];
  traffic_sources: TrafficSource[];
  visitor_history: DailyVisitors[];
  funnel: FunnelStep[];
}

export interface RevenueResponse {
  mrr: number;
  arr: number;
  total_revenue: number;
  paying_users: number;
  churn_rate: number;
  ltv: number;
  mrr_history: { month: string; amount: number }[];
  plan_breakdown: PlanRevenue[];
  failed_payments: FailedPayment[];
}

export interface AIHealthResponse {
  total_queries: number;
  answered_correctly: number;
  accuracy: number;
  accuracy_trend: number;
  unanswered_questions: UnansweredQuestion[];
  accuracy_history: DailyAccuracy[];
  sentiment_breakdown: SentimentBreakdown;
}

export interface SystemHealthResponse {
  services: { name: string; status: string; latency_ms: number }[];
  error_rate: number;
  p50_latency: number;
  p95_latency: number;
  p99_latency: number;
}

export interface AlertsResponse {
  alerts: { id: string; type: string; title: string; description: string; severity: string; created_at: string }[];
}

export interface ActivityResponse {
  events: LiveFeedEvent[];
}
