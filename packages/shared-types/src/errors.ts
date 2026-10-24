// Standard Go API Error Structure matching Section 16 of Master Directive

export interface ApiErrorDetail {
  code: string;
  message: string;
  request_id: string;
  field?: string;
  details?: Record<string, unknown>;
}

export interface ApiErrorResponse {
  error: ApiErrorDetail;
}

export enum ApiErrorCode {
  UNAUTHENTICATED = 'UNAUTHENTICATED',
  FORBIDDEN_RESOURCE = 'FORBIDDEN_RESOURCE',
  NOT_FOUND = 'NOT_FOUND',
  INVALID_ARGUMENT = 'INVALID_ARGUMENT',
  CONFLICT = 'CONFLICT',
  RATE_LIMITED = 'RATE_LIMITED',
  INTERNAL_ERROR = 'INTERNAL_ERROR',
  UPLOAD_EXPIRED = 'UPLOAD_EXPIRED',
  QUOTA_EXCEEDED = 'QUOTA_EXCEEDED',
  POLICY_VIOLATION = 'POLICY_VIOLATION',
  PAYMENT_REQUIRED = 'PAYMENT_REQUIRED'
}

