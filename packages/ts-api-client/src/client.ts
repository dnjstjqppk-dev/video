import type {
  User,
  UserProfile,
  Channel,
  Video,
  PlaybackInfo,
  Comment,
  VideoCaption,
  Playlist,
  PlaylistItem,
  LiveStream,
  AdCampaign,
  AdCreative,
  AdDecisionRequest,
  AdDecisionResponse,
  AdEventPayload,
  CopyrightClaim,
  Payout,
  ModerationCase,
  PartnerContract,
  ApiErrorResponse,
  AnalyticsPlaybackEvent,
  StudioEditDecisionList,
  StudioEditJob
} from '@gvp/shared-types';
import {
  toCamelDeep,
  mapUser,
  mapUserProfile,
  mapChannel,
  mapVideo,
  mapPlayback,
  mapComment,
  mapVideoCaption,
  mapLiveStream,
  mapAdCreative,
  mapAdDecision,
  mapCampaign,
  mapCopyrightClaim,
  mapPayout,
  mapModerationCase,
  mapPartnerContract
} from './normalize';

export interface ApiClientConfig {
  baseUrl: string;
  getToken?: () => string | null | Promise<string | null>;
  onUnauthorized?: () => void;
}

export class ApiError extends Error {
  code: string;
  requestId: string;
  field?: string;
  status: number;

  constructor(status: number, code: string, message: string, requestId: string, field?: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.requestId = requestId;
    this.field = field;
  }
}

function readCsrfCookie(): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)gvp_csrf=([^;]*)/);
  const value = match?.[1];
  return value ? decodeURIComponent(value) : null;
}

export class ApiClient {
  private baseUrl: string;
  private getToken?: () => string | null | Promise<string | null>;
  private onUnauthorized?: () => void;

  constructor(config: ApiClientConfig) {
    this.baseUrl = config.baseUrl.replace(/\/$/, '');
    this.getToken = config.getToken;
    this.onUnauthorized = config.onUnauthorized;
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const headers = new Headers(options.headers || {});

    if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
      headers.set('Content-Type', 'application/json');
    }

    if (!headers.has('X-Request-Id')) {
      headers.set('X-Request-Id', `req_${Math.random().toString(36).substring(2, 11)}_${Date.now()}`);
    }

    if (this.getToken) {
      const token = await this.getToken();
      if (token && !headers.has('Authorization')) {
        headers.set('Authorization', `Bearer ${token}`);
      }
    }

    // Double-submit CSRF token for mutations through the Nuxt session proxy.
    const method = (options.method || 'GET').toUpperCase();
    if (method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS' && !headers.has('X-CSRF-Token')) {
      const csrf = readCsrfCookie();
      if (csrf) {
        headers.set('X-CSRF-Token', csrf);
      }
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      let errorBody: ApiErrorResponse | null = null;
      try {
        errorBody = await response.json();
      } catch {
        // Not a JSON response
      }

      if (response.status === 401 && this.onUnauthorized) {
        this.onUnauthorized();
      }

      const code = errorBody?.error?.code || `HTTP_${response.status}`;
      const message = errorBody?.error?.message || response.statusText || 'An unexpected error occurred';
      const requestId = errorBody?.error?.request_id || headers.get('X-Request-Id') || '';
      const field = errorBody?.error?.field;

      throw new ApiError(response.status, code, message, requestId, field);
    }

    if (response.status === 204) {
      return {} as T;
    }

    return response.json();
  }

  // --- Auth & User ---
  async getMe(): Promise<UserProfile> {
    const raw = await this.request<unknown>('/v1/me');
    return mapUserProfile(raw);
  }

  async updateMe(data: Partial<UserProfile>): Promise<UserProfile> {
    const raw = await this.request<unknown>('/v1/me', {
      method: 'PATCH',
      body: JSON.stringify({
        display_name: data.displayName,
        bio: data.bio
      }),
    });
    return mapUserProfile(raw);
  }

  async getSessions(): Promise<Array<{ id: string; ip: string; userAgent: string; lastActive: string; isCurrent: boolean }>> {
    const raw = await this.request<unknown>('/v1/me/sessions');
    return toCamelDeep(raw) as Array<{ id: string; ip: string; userAgent: string; lastActive: string; isCurrent: boolean }>;
  }

  async revokeSession(id: string): Promise<void> {
    return this.request(`/v1/me/sessions/${id}`, { method: 'DELETE' });
  }

  async getWatchHistory(page = 1, limit = 20): Promise<{ items: Video[]; total: number }> {
    const raw = await this.request<{ items: unknown[]; total: number }>(`/v1/me/history?page=${page}&limit=${limit}`);
    return { items: (raw.items || []).map(mapVideo), total: raw.total || 0 };
  }

  async clearWatchHistory(): Promise<void> {
    return this.request('/v1/me/history', { method: 'DELETE' });
  }

  // --- Channels ---
  async getChannel(id: string): Promise<Channel> {
    const raw = await this.request<unknown>(`/v1/channels/${id}`);
    return mapChannel(raw);
  }

  async createChannel(data: { name: string; handle: string; description?: string }): Promise<Channel> {
    const raw = await this.request<unknown>('/v1/channels', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return mapChannel(raw);
  }

  async updateChannel(id: string, data: Partial<Channel>): Promise<Channel> {
    const raw = await this.request<unknown>(`/v1/channels/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
    return mapChannel(raw);
  }

  async getChannelVideos(channelId: string, page = 1, limit = 20): Promise<{ items: Video[]; total: number }> {
    const raw = await this.request<{ items: unknown[]; total: number }>(`/v1/channels/${channelId}/videos?page=${page}&limit=${limit}`);
    return { items: (raw.items || []).map(mapVideo), total: raw.total || 0 };
  }

  async subscribeChannel(channelId: string): Promise<{ subscribed: boolean; subscriberCount: number }> {
    const raw = await this.request<unknown>(`/v1/channels/${channelId}/subscribe`, { method: 'POST' });
    return toCamelDeep(raw) as { subscribed: boolean; subscriberCount: number };
  }

  async unsubscribeChannel(channelId: string): Promise<{ subscribed: boolean; subscriberCount: number }> {
    const raw = await this.request<unknown>(`/v1/channels/${channelId}/subscribe`, { method: 'DELETE' });
    return toCamelDeep(raw) as { subscribed: boolean; subscriberCount: number };
  }

  // --- Uploads ---
  async createUploadSession(data: {
    title: string;
    description?: string;
    fileSizeBytes: number;
    mimeType: string;
    videoType?: string;
  }): Promise<{
    uploadId: string;
    videoId: string;
    partSize: number;
    totalParts: number;
    key: string;
  }> {
    return this.request('/v1/uploads', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getUploadPartUrls(uploadId: string, partNumber: number): Promise<{ uploadUrl: string; partNumber: number }> {
    return this.request(`/v1/uploads/${uploadId}/parts`, {
      method: 'POST',
      body: JSON.stringify({ partNumber }),
    });
  }

  async completeUpload(uploadId: string, parts: Array<{ partNumber: number; etag: string }>): Promise<{ videoId: string; status: string }> {
    const raw = await this.request<unknown>(`/v1/uploads/${uploadId}/complete`, {
      method: 'POST',
      body: JSON.stringify({ parts }),
    });
    return toCamelDeep(raw) as { videoId: string; status: string };
  }

  // --- Videos & Playback ---
  async getVideo(id: string): Promise<Video> {
    const raw = await this.request<unknown>(`/v1/videos/${id}`);
    return mapVideo(raw);
  }

  async getPlaybackInfo(id: string): Promise<PlaybackInfo> {
    const raw = await this.request<unknown>(`/v1/videos/${id}/playback`);
    return mapPlayback(raw);
  }

  async updateVideo(id: string, data: Partial<Video>): Promise<Video> {
    const raw = await this.request<unknown>(`/v1/videos/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
    return mapVideo(raw);
  }

  async deleteVideo(id: string): Promise<void> {
    return this.request(`/v1/videos/${id}`, { method: 'DELETE' });
  }

  async likeVideo(id: string): Promise<{ liked: boolean; likeCount: number }> {
    const raw = await this.request<unknown>(`/v1/videos/${id}/like`, { method: 'POST' });
    return toCamelDeep(raw) as { liked: boolean; likeCount: number };
  }

  async unlikeVideo(id: string): Promise<{ liked: boolean; likeCount: number }> {
    const raw = await this.request<unknown>(`/v1/videos/${id}/like`, { method: 'DELETE' });
    return toCamelDeep(raw) as { liked: boolean; likeCount: number };
  }

  // --- Search & Feeds ---
  async search(query: string, options: { page?: number; limit?: number; type?: string; sort?: string } = {}): Promise<{
    videos: Video[];
    channels: Channel[];
    total: number;
  }> {
    const params = new URLSearchParams({
      q: query,
      page: String(options.page || 1),
      limit: String(options.limit || 20),
      ...(options.type ? { type: options.type } : {}),
      ...(options.sort ? { sort: options.sort } : {})
    });
    const raw = await this.request<{ videos: unknown[]; channels: unknown[]; total: number }>(`/v1/search?${params.toString()}`);
    return {
      videos: (raw.videos || []).map(mapVideo),
      channels: (raw.channels || []).map(mapChannel),
      total: raw.total || 0
    };
  }

  async getSearchSuggestions(query: string): Promise<string[]> {
    return this.request(`/v1/search/suggestions?q=${encodeURIComponent(query)}`);
  }

  async getHomeFeed(page = 1, limit = 24, category?: string): Promise<{ items: Video[]; nextCursor?: string }> {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (category && category !== 'all') params.set('category', category);
    const raw = await this.request<{ items: unknown[]; nextCursor?: string }>(`/v1/feed/home?${params.toString()}`);
    return { items: (raw.items || []).map(mapVideo), nextCursor: raw.nextCursor };
  }

  async getShortsFeed(page = 1, limit = 10): Promise<{ items: Video[]; nextCursor?: string }> {
    const raw = await this.request<{ items: unknown[]; nextCursor?: string }>(`/v1/feed/shorts?page=${page}&limit=${limit}`);
    return { items: (raw.items || []).map(mapVideo), nextCursor: raw.nextCursor };
  }

  async getSubscriptionsFeed(page = 1, limit = 24): Promise<{ items: Video[]; total: number }> {
    const raw = await this.request<{ items: unknown[]; total: number }>(`/v1/feed/subscriptions?page=${page}&limit=${limit}`);
    return { items: (raw.items || []).map(mapVideo), total: raw.total || 0 };
  }

  async sendFeedFeedback(videoId: string, action: 'not_interested' | 'dont_recommend_channel'): Promise<void> {
    return this.request('/v1/feed/feedback', {
      method: 'POST',
      body: JSON.stringify({ videoId, action }),
    });
  }

  // --- Comments ---
  async getVideoComments(videoId: string, page = 1, limit = 20, sort: 'top' | 'newest' = 'top'): Promise<{ items: Comment[]; total: number }> {
    const raw = await this.request<{ items: unknown[]; total: number }>(`/v1/videos/${videoId}/comments?page=${page}&limit=${limit}&sort=${sort}`);
    return { items: (raw.items || []).map(mapComment), total: raw.total || 0 };
  }

  async addComment(videoId: string, content: string, parentId?: string): Promise<Comment> {
    const raw = await this.request<unknown>(`/v1/videos/${videoId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ content, parent_id: parentId }),
    });
    return mapComment(raw);
  }

  async deleteComment(commentId: string): Promise<void> {
    return this.request(`/v1/comments/${commentId}`, { method: 'DELETE' });
  }

  async likeComment(commentId: string): Promise<{ liked: boolean; likeCount: number }> {
    const raw = await this.request<{ like_count?: number; likeCount?: number }>(`/v1/comments/${commentId}/like`, { method: 'POST' });
    return { liked: true, likeCount: raw.like_count ?? raw.likeCount ?? 0 };
  }

  async heartComment(commentId: string): Promise<{ isHearted: boolean }> {
    const raw = await this.request<{ is_hearted?: boolean; isHearted?: boolean }>(`/v1/comments/${commentId}/heart`, { method: 'POST' });
    return { isHearted: Boolean(raw.is_hearted ?? raw.isHearted) };
  }

  async pinComment(commentId: string): Promise<{ isPinned: boolean }> {
    const raw = await this.request<{ is_pinned?: boolean; isPinned?: boolean }>(`/v1/comments/${commentId}/pin`, { method: 'POST' });
    return { isPinned: Boolean(raw.is_pinned ?? raw.isPinned) };
  }

  // --- Playlists ---
  async createPlaylist(data: { title: string; description?: string; visibility?: string }): Promise<Playlist> {
    return this.request('/v1/playlists', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getPlaylist(id: string): Promise<{ playlist: Playlist; items: PlaylistItem[] }> {
    return this.request(`/v1/playlists/${id}`);
  }

  async addPlaylistItem(playlistId: string, videoId: string): Promise<PlaylistItem> {
    return this.request(`/v1/playlists/${playlistId}/items`, {
      method: 'POST',
      body: JSON.stringify({ videoId }),
    });
  }

  // --- Live Streams ---
  async createLiveStream(data: { title: string; description?: string }): Promise<LiveStream> {
    const raw = await this.request<unknown>('/v1/live-streams', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return mapLiveStream(raw);
  }

  async getLiveStream(id: string): Promise<LiveStream> {
    const raw = await this.request<unknown>(`/v1/live-streams/${id}`);
    return mapLiveStream(raw);
  }

  async startLiveStream(id: string): Promise<LiveStream> {
    const raw = await this.request<unknown>(`/v1/live-streams/${id}/start`, { method: 'POST' });
    return mapLiveStream(raw);
  }

  async stopLiveStream(id: string): Promise<LiveStream> {
    const raw = await this.request<unknown>(`/v1/live-streams/${id}/stop`, { method: 'POST' });
    return mapLiveStream(raw);
  }

  async rotateLiveStreamKey(id: string): Promise<{ streamKey: string }> {
    return this.request(`/v1/live-streams/${id}/rotate-key`, { method: 'POST' });
  }

  // --- Analytics & Events ---
  async recordAnalyticsEvent(event: AnalyticsPlaybackEvent): Promise<{ status: string }> {
    return this.request('/v1/events/analytics', {
      method: 'POST',
      body: JSON.stringify(event),
    });
  }

  // --- Studio ---
  async getStudioDashboard(): Promise<{
    channel: Channel;
    analytics: { views: number; watchTimeHours: number; subscribers: number; estimatedRevenue: number };
    recentVideos: Video[];
    recentComments: Comment[];
  }> {
    const raw = await this.request<{ channel: unknown; analytics: { views: number; watchTimeHours: number; subscribers: number; estimatedRevenue: number }; recentVideos?: unknown[]; recentComments?: unknown[] }>('/v1/studio/dashboard');
    return {
      channel: mapChannel(raw.channel),
      analytics: raw.analytics,
      recentVideos: (raw.recentVideos || []).map(mapVideo),
      recentComments: (raw.recentComments || []).map(mapComment),
    };
  }

  async getStudioVideos(page = 1, limit = 20, status?: string): Promise<{ items: Video[]; total: number }> {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (status) params.set('status', status);
    const raw = await this.request<{ items: unknown[]; total: number }>(`/v1/studio/videos?${params.toString()}`);
    return { items: (raw.items || []).map(mapVideo), total: raw.total || 0 };
  }

  async listVideoCaptions(videoId: string): Promise<VideoCaption[]> {
    const raw = await this.request<{ items: unknown[]; total: number }>(`/v1/videos/${videoId}/captions`);
    return (raw.items || []).map((c, i) => mapVideoCaption(c, i));
  }

  async addVideoCaption(videoId: string, data: { language: string; label?: string; vttUrl: string }): Promise<VideoCaption> {
    const raw = await this.request<unknown>(`/v1/videos/${videoId}/captions`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return mapVideoCaption(raw, 0);
  }

  async deleteVideoCaption(videoId: string, captionId: string): Promise<void> {
    return this.request(`/v1/videos/${videoId}/captions/${captionId}`, { method: 'DELETE' });
  }

  async getStudioComments(page = 1, limit = 20): Promise<{ items: Comment[]; total: number }> {
    const raw = await this.request<{ items: unknown[]; total: number }>(`/v1/studio/comments?page=${page}&limit=${limit}`);
    return { items: (raw.items || []).map(mapComment), total: raw.total || 0 };
  }

  async getStudioAnalytics(timeRange = '28d'): Promise<{
    viewsTimeSeries: Array<{ date: string; value: number }>;
    watchTimeTimeSeries: Array<{ date: string; value: number }>;
    subscriberTimeSeries: Array<{ date: string; value: number }>;
    revenueTimeSeries: Array<{ date: string; value: number }>;
  }> {
    return this.request(`/v1/studio/analytics?range=${timeRange}`);
  }

  // --- Partners & Advertising Engine ---
  async getAdDecision(data: AdDecisionRequest): Promise<AdDecisionResponse> {
    const raw = await this.request<unknown>('/v1/ads/decision', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return mapAdDecision(raw);
  }

  async recordAdEvent(data: AdEventPayload): Promise<{ status: string }> {
    return this.request('/v1/ads/events', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getAdDeliveryDiagnostics(): Promise<any> {
    return this.request('/v1/admin/ads/delivery-diagnostics');
  }

  async getAdPreview(data: { slot_id: string; category?: string }): Promise<AdDecisionResponse> {
    const raw = await this.request<unknown>('/v1/ads/preview', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return mapAdDecision(raw);
  }

  async getPartnerOverview(): Promise<{
    organization: { id: string; name: string; type: string };
    campaignCount: number;
    activeChannelsCount: number;
    pendingSettlementAmount: number;
  }> {
    return this.request('/v1/partners/me');
  }

  async getPartnerCampaigns(): Promise<AdCampaign[]> {
    const raw = await this.request<unknown[]>('/v1/partners/campaigns');
    return (raw || []).map(mapCampaign);
  }

  async createPartnerCampaign(data: Partial<AdCampaign>): Promise<AdCampaign> {
    const raw = await this.request<unknown>('/v1/partners/campaigns', {
      method: 'POST',
      body: JSON.stringify({
        name: data.name,
        format: data.format,
        totalBudget: Math.round((data.totalBudget || 0) * 100),
        dailyBudget: Math.round((data.dailyBudget || 0) * 100),
        cpmBidCents: Math.round((data.cpmBid || 0) * 100),
        cpcBidCents: Math.round((data.cpcBid || 0) * 100),
        categories: data.targetCategories || []
      }),
    });
    return mapCampaign(raw);
  }

  async getPartnerContracts(): Promise<PartnerContract[]> {
    const raw = await this.request<{ items: unknown[]; total: number }>('/v1/partners/contracts');
    return (raw.items || []).map(mapPartnerContract);
  }

  async revokePartnerContract(id: string, reason: string): Promise<PartnerContract> {
    const raw = await this.request<unknown>(`/v1/partners/contracts/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ reason }),
    });
    return mapPartnerContract(raw);
  }

  async getStudioClaims(): Promise<CopyrightClaim[]> {
    const raw = await this.request<{ items: unknown[]; total: number }>('/v1/studio/claims');
    return (raw.items || []).map(mapCopyrightClaim);
  }

  async disputeCopyrightClaim(id: string, reason: string, explanation?: string): Promise<CopyrightClaim> {
    const raw = await this.request<unknown>(`/v1/partners/rights/claims/${id}/dispute`, {
      method: 'POST',
      body: JSON.stringify({ reason, explanation: explanation || '' }),
    });
    return mapCopyrightClaim(raw);
  }

  async getPartnerClaims(): Promise<CopyrightClaim[]> {
    const raw = await this.request<unknown[]>('/v1/partners/rights/claims');
    return (raw || []).map(mapCopyrightClaim);
  }

  async createCopyrightClaim(data: { videoId: string; assetId: string; action: string }): Promise<CopyrightClaim> {
    const raw = await this.request<unknown>('/v1/partners/rights/claims', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return mapCopyrightClaim(raw);
  }

  async getPartnerSettlements(): Promise<Payout[]> {
    const raw = await this.request<unknown[]>('/v1/partners/settlements');
    return (raw || []).map(mapPayout);
  }

  // --- Admin ---
  async getServiceHealth(): Promise<{
    alive: { status: string; version: string; timestamp: string } | null;
    ready: { status: string; ready: boolean; timestamp: string; subsystems: Record<string, string> } | null;
  }> {
    const [alive, ready] = await Promise.all([
      this.request<{ status: string; version: string; timestamp: string }>('/healthz').catch(() => null),
      this.request<{ status: string; ready: boolean; timestamp: string; subsystems: Record<string, string> }>('/ready').catch(() => null),
    ]);
    return { alive, ready };
  }

  async getAdminDashboard(): Promise<{
    metrics: {
      totalUsers: number;
      activeUsersToday: number;
      totalVideos: number;
      liveStreamsActive: number;
      totalAdRevenueKrw: number;
      systemHealth: string;
    };
    recentAuditLogs: Array<{ id: string; action: string; actor: string; timestamp: string }>;
  }> {
    return this.request('/v1/admin/dashboard');
  }

  async getAdminUsers(page = 1, limit = 20): Promise<{ items: User[]; total: number }> {
    const raw = await this.request<{ items: unknown[]; total: number }>(`/v1/admin/users?page=${page}&limit=${limit}`);
    return { items: (raw.items || []).map(mapUser), total: raw.total || 0 };
  }

  async updateAdminUserStatus(userId: string, status: 'active' | 'suspended' | 'banned', reason: string): Promise<User> {
    const raw = await this.request<unknown>(`/v1/admin/users/${userId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status, reason }),
    });
    return mapUser(raw);
  }

  async reportContent(data: { targetType: 'video' | 'comment' | 'channel'; targetId: string; reason: string; policyCategory?: string; details?: string }): Promise<ModerationCase> {
    const raw = await this.request<unknown>('/v1/moderation/reports', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return mapModerationCase(raw);
  }

  async getAdminModerationQueue(): Promise<ModerationCase[]> {
    const raw = await this.request<unknown[]>('/v1/admin/moderation');
    return (raw || []).map(mapModerationCase);
  }

  async resolveModerationCase(caseId: string, action: 'approve' | 'reject' | 'delete', notes: string): Promise<ModerationCase> {
    const raw = await this.request<unknown>(`/v1/admin/moderation/${caseId}/resolve`, {
      method: 'POST',
      body: JSON.stringify({ action, notes }),
    });
    return mapModerationCase(raw);
  }

  async getAdminPolicies(): Promise<{
    config: {
      version: string;
      max_upload_size_gb: number;
      max_video_duration_hours: number;
      creator_revenue_share_pct: number;
      updated_by: string;
      updated_at: string;
    };
    flags: Array<{
      key: string;
      label: string;
      description: string;
      enabled: boolean;
      updated_by: string;
      updated_at: string;
    }>;
  }> {
    return this.request('/v1/admin/policies');
  }

  async updateAdminPolicyConfig(data: {
    max_upload_size_gb?: number;
    max_video_duration_hours?: number;
    creator_revenue_share_pct?: number;
  }): Promise<{
    version: string;
    max_upload_size_gb: number;
    max_video_duration_hours: number;
    creator_revenue_share_pct: number;
  }> {
    return this.request('/v1/admin/policies', {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async updateAdminFlag(key: string, enabled: boolean): Promise<{ key: string; enabled: boolean }> {
    return this.request(`/v1/admin/policies/flags/${encodeURIComponent(key)}`, {
      method: 'PATCH',
      body: JSON.stringify({ enabled }),
    });
  }

  async getPartnerEarnings(): Promise<{
    currentMonthEstimated: number;
    adRevenue: number;
    membershipRevenue: number;
    superChatRevenue: number;
    currency: string;
  }> {
    return this.request('/v1/partners/settlements/earnings-summary');
  }

  async requestPayout(channelId?: string): Promise<{ payout_id: string; status: string; message: string }> {
    return this.request('/v1/partners/settlements/payout', {
      method: 'POST',
      body: JSON.stringify(channelId ? { channelId } : {}),
    });
  }

  async getStudioSettlements(): Promise<Payout[]> {
    const raw = await this.request<unknown[]>('/v1/studio/settlements');
    return (raw || []).map(mapPayout);
  }

  async getStudioEarnings(): Promise<{
    currentMonthEstimated: number;
    adRevenue: number;
    membershipRevenue: number;
    superChatRevenue: number;
    currency: string;
  }> {
    return this.request('/v1/studio/earnings');
  }

  async createStudioEditJob(videoId: string, edl: StudioEditDecisionList): Promise<StudioEditJob> {
    return this.request(`/v1/studio/videos/${videoId}/edit-jobs`, {
      method: 'POST',
      body: JSON.stringify(edl)
    });
  }

  async getStudioEditJob(videoId: string, jobId: string): Promise<StudioEditJob> {
    return this.request(`/v1/studio/videos/${videoId}/edit-jobs/${jobId}`);
  }

  async setupMfa(): Promise<{ secret: string; otpauth_url: string }> {
    return this.request('/v1/me/mfa/setup', { method: 'POST' });
  }

  async verifyMfa(code: string): Promise<{ verified: boolean }> {
    return this.request('/v1/me/mfa/verify', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
  }

  async enableMfa(code: string): Promise<{ enabled: boolean }> {
    return this.request('/v1/me/mfa/enable', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
  }

  async disableMfa(): Promise<{ enabled: boolean }> {
    return this.request('/v1/me/mfa/disable', { method: 'POST' });
  }

  async getAdminAuditLogs(page = 1, limit = 50): Promise<{ items: Array<{ id: string; actorId: string; action: string; target: string; ip: string; timestamp: string }>; total: number }> {
    const raw = await this.request<{ items: Record<string, any>[]; total: number }>(`/v1/admin/audit-logs?page=${page}&limit=${limit}`);
    const items = (raw.items || []).map((entry, index) => {
      const e = toCamelDeep<Record<string, any>>(entry);
      return {
        id: typeof e.id === 'string' ? e.id : `audit_${index}_${e.timestamp || ''}`,
        actorId: typeof e.actorId === 'string' ? e.actorId : '',
        action: typeof e.action === 'string' ? e.action : '',
        target: typeof e.target === 'string' ? e.target : '',
        ip: typeof e.ip === 'string' ? e.ip : '',
        timestamp: typeof e.timestamp === 'string' ? e.timestamp : ''
      };
    });
    return { items, total: raw.total || 0 };
  }

  async getAdminFinanceSummary(): Promise<{
    balanced: boolean;
    debitTotalCents: number;
    creditTotalCents: number;
    debitByAccount: Record<string, number>;
    creditByAccount: Record<string, number>;
    transactionsByStatus: Record<string, number>;
    succeededCents: number;
    creatorPayableCents: number;
    ledgerEntries: number;
    batchId: string;
    generatedAt: string;
  }> {
    return this.request('/v1/admin/finance/summary');
  }

  async getAdminFinanceSettlements(): Promise<Payout[]> {
    const raw = await this.request<unknown[]>('/v1/admin/finance/settlements');
    return (raw || []).map(mapPayout);
  }

  async approveSettlementBatch(batchId: string, totpCode: string): Promise<{
    approvalId: string;
    batchId: string;
    approvedBy: string;
    debitTotal: number;
    creditTotal: number;
    status: string;
  }> {
    return this.request('/v1/admin/finance/batches/approve', {
      method: 'POST',
      body: JSON.stringify({ batchId, totpCode }),
    });
  }
}

export function createApiClient(config: ApiClientConfig): ApiClient {
  return new ApiClient(config);
}
