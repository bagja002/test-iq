export type Role = "USER" | "ADMIN"
export type AttemptStatus = "IN_PROGRESS" | "SUBMITTED" | "EXPIRED"
export type AttemptSectionStatus = "NOT_STARTED" | "IN_PROGRESS" | "SUBMITTED" | "EXPIRED"
export type QuestionStatus = "DRAFT" | "PUBLISHED" | "ARCHIVED"
export type UserStatus = "ACTIVE" | "INACTIVE"
export type AccountType = "FREE" | "PRO" | "MAX"
export type QuestionIndex = "VCI" | "PRI" | "WMI" | "PSI" | "SKB"
export type TestType = "IQ" | "SKB"

export interface ApiError {
  message: string
  details?: string
}

export interface SessionUser {
  id: number
  name: string
  position: string
  email: string
  role: Role
  status: UserStatus
  accountType: AccountType
}

export interface SessionResponse {
  authenticated: boolean
  user: SessionUser | null
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  name: string
  position: string
  email: string
  password: string
}

export interface LoginResponse {
  message: string
  user: SessionUser
}

export interface QuestionOptionInput {
  key: string
  content: string
  mediaUrl?: string
  mediaAlt?: string
  isCorrect?: boolean
}

export interface QuestionPayload {
  prompt: string
  promptMediaUrl?: string
  promptMediaAlt?: string
  difficulty?: string
  questionIndex: QuestionIndex
  subtestCode: string
  status: QuestionStatus
  options: QuestionOptionInput[]
}

export interface QuestionSummary {
  id: number
  prompt: string
  promptMediaUrl: string | null
  promptMediaAlt: string | null
  difficulty: string
  questionIndex: QuestionIndex | null
  questionIndexLabel: string | null
  subtestCode: string | null
  subtestLabel: string | null
  status: QuestionStatus
  hasOptionMedia: boolean
  optionCount: number
  updatedAt: string
}

export interface TestConfigResponse {
  id: number
  title: string
  testType: TestType
  roomCode: string | null
  roomLabel: string | null
  durationMinutes: number
  questionCount: number
  isActive: boolean
  updatedAt: string
}

export interface AdminQuestionStats {
  total: number
  published: number
  draft: number
  archived: number
  visual: number
}

export interface AdminUserStats {
  total: number
  active: number
  inactive: number
  adminCount: number
  participantCount: number
}

export interface AdminAttemptStats {
  inProgress: number
  submitted: number
  expired: number
}

export interface AdminResultStats {
  submissionCount: number
  averagePercentage: number
  highestEstimatedIq: number
}

export interface AdminQuestionHealth {
  code: QuestionIndex
  label: string
  published: number
  total: number
}

export interface AdminConfigHealth {
  id: number
  title: string
  testType: TestType
  roomCode: string | null
  roomLabel: string | null
  durationMinutes: number
  questionCount: number
  publishedQuestionCount: number
  canStartAttempt: boolean
  readinessMessage: string
  questionHealth: AdminQuestionHealth[]
}

export interface AdminOverview {
  questionStats: AdminQuestionStats
  userStats: AdminUserStats
  attemptStats: AdminAttemptStats
  resultStats: AdminResultStats
  questionHealth: AdminQuestionHealth[]
  activeConfig: AdminConfigHealth | null
}

export interface AttemptQuestionOption {
  key: string
  content: string
  mediaUrl: string
  mediaAlt: string
}

export interface AttemptQuestion {
  id: number
  questionId: number
  orderNo: number
  questionIndex: QuestionIndex | null
  questionIndexLabel: string | null
  subtestCode: string | null
  subtestLabel: string | null
  prompt: string
  promptMediaUrl: string
  promptMediaAlt: string
  options: AttemptQuestionOption[]
  selectedOptionKey: string | null
}

export interface AttemptSection {
  id: number
  code: QuestionIndex
  label: string
  orderNo: number
  status: AttemptSectionStatus
  durationMinutes: number
  questionCount: number
  answeredCount: number
  startedAt: string | null
  expiresAt: string | null
  submittedAt: string | null
}

export interface AttemptDetail {
  id: number
  testType: TestType
  roomCode: string | null
  roomLabel: string | null
  status: AttemptStatus
  startedAt: string
  expiresAt: string
  submittedAt: string | null
  durationMinutes: number
  rawScore: number | null
  totalQuestions: number
  percentage: number | null
  sections: AttemptSection[]
  questions: AttemptQuestion[]
}

export interface AttemptSummary {
  id: number
  testType: TestType
  roomCode: string | null
  roomLabel: string | null
  status: AttemptStatus
  startedAt: string
  expiresAt: string
  submittedAt: string | null
  rawScore: number | null
  totalQuestions: number
  percentage: number | null
}

export interface IndexScore {
  code: QuestionIndex
  label: string
  correct: number
  total: number
  percentage: number
}

export interface AttemptResult extends AttemptSummary {
  estimatedIq: number | null
  skbScore: number | null
  classificationLabel: string | null
  indexScores: IndexScore[]
}

export interface ResultListResponse {
  results: AttemptResult[]
}

export interface SaveAnswersRequest {
  answers: Array<{
    attemptQuestionId: number
    selectedOptionKey: string
  }>
}

export interface ResultResponse {
  attempt: AttemptResult | null
}

export interface UserSummary {
  id: number
  name: string
  position: string
  email: string
  role: Role
  status: UserStatus
  accountType: AccountType
  createdAt: string
}

export interface AdminResultSummary {
  attemptId: number
  userId: number
  userName: string
  email: string
  rawScore: number
  totalQuestions: number
  percentage: number
  submittedAt: string
  durationMinutes: number
  estimatedIq: number
  classificationLabel: string
}
