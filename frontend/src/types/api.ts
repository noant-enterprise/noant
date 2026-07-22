import type {
  Conversation, Message, TrendData, IntentData, PeakHourData,
  ChannelDistribution, Integration, TeamMember, APIKey, Category,
  UnknownQuestion, HandoffsResponse,
} from './index'

export type { HandoffsResponse }

export interface ConversationListResponse {
  conversations: Conversation[]
  has_more: boolean
}

export interface ConversationMessagesResponse {
  messages: Message[]
  has_more: boolean
}

export interface DirectChatResponse {
  conversation: { id: string; [key: string]: unknown }
}

export interface AnalyticsOverview {
  conversations_today: number
  resolved_today: number
  avg_response_time: number
  satisfaction: number
  total_conversations: number
}

export interface TrendsResponse {
  trends: TrendData[]
}

export interface InsightsResponse {
  top_intents: IntentData[]
  peak_hours: PeakHourData[]
}

export interface ChannelAnalyticsResponse {
  distribution: ChannelDistribution
}

export interface CSATResponse {
  avg_score: number
  total_ratings: number
  distribution: Record<string, number>
  trend: Array<{ date: string; avg_score: number }>
}

export interface UnknownQuestionsStatsResponse {
  by_status: { pending: number; trained: number; ignored: number }
  total: number
}

export interface PopularQuestionsResponse {
  questions: Array<{ question: string; count: number }>
}

export interface MessageTrendResponse {
  trends: Array<{ date: string; messages: number }>
}

export interface IntegrationsResponse {
  integrations: Integration[]
}

export interface TrainingCategoriesResponse {
  categories: Category[]
}

export interface TrainingUnknownQuestionsResponse {
  questions: UnknownQuestion[]
}

export interface ProfileResponse {
  first_name: string
  last_name: string
  email: string
  company_name: string
  phone: string
}

export interface TeamResponse {
  members: TeamMember[]
}

export interface APIKeysResponse {
  api_keys: APIKey[]
}
