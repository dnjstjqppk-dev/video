import type {
  User,
  UserProfile,
  Channel,
  Video,
  PlaybackInfo,
  Comment,
  Playlist,
  PlaylistItem,
  LiveStream,
  LiveChatMessage,
  AdCampaign,
  AdCreative,
  AdDecisionRequest,
  AdDecisionResponse,
  AdEventPayload,
  CopyrightClaim,
  Payout,
  ModerationCase,
  ApiErrorResponse,
  AnalyticsPlaybackEvent
} from '@gvp/shared-types';

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
    return this.request<UserProfile>('/v1/me');
  }

  async updateMe(data: Partial<UserProfile>): Promise<UserProfile> {
    return this.request<UserProfile>('/v1/me', {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async getSessions(): Promise<Array<{ id: string; ip: string; userAgent: string; lastActive: string; isCurrent: boolean }>> {
    return this.request('/v1/me/sessions');
  }

  async revokeSession(id: string): Promise<void> {
    return this.request(`/v1/me/sessions/${id}`, { method: 'DELETE' });
  }

  async getWatchHistory(page = 1, limit = 20): Promise<{ items: Video[]; total: number }> {
    return this.request(`/v1/me/history?page=${page}&limit=${limit}`);
  }

  async clearWatchHistory(): Promise<void> {
    return this.request('/v1/me/history', { method: 'DELETE' });
  }

  // --- Channels ---
  async getChannel(id: string): Promise<Channel> {
    return this.request<Channel>(`/v1/channels/${id}`);
  }

  async createChannel(data: { name: string; handle: string; description?: string }): Promise<Channel> {
    return this.request<Channel>('/v1/channels', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateChannel(id: string, data: Partial<Channel>): Promise<Channel> {
    return this.request<Channel>(`/v1/channels/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async getChannelVideos(channelId: string, page = 1, limit = 20): Promise<{ items: Video[]; total: number }> {
    return this.request(`/v1/channels/${channelId}/videos?page=${page}&limit=${limit}`);
  }

  async subscribeChannel(channelId: string): Promise<{ subscribed: boolean; subscriberCount: number }> {
    return this.request(`/v1/channels/${channelId}/subscribe`, { method: 'POST' });
  }

  async unsubscribeChannel(channelId: string): Promise<{ subscribed: boolean; subscriberCount: number }> {
    return this.request(`/v1/channels/${channelId}/subscribe`, { method: 'DELETE' });
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
    return this.request(`/v1/uploads/${uploadId}/complete`, {
      method: 'POST',
      body: JSON.stringify({ parts }),
    });
  }

  // --- Videos & Playback ---
  async getVideo(id: string): Promise<Video> {
    return this.request<Video>(`/v1/videos/${id}`);
  }

  async getPlaybackInfo(id: string): Promise<PlaybackInfo> {
    return this.request<PlaybackInfo>(`/v1/videos/${id}/playback`);
  }

  async updateVideo(id: string, data: Partial<Video>): Promise<Video> {
    return this.request<Video>(`/v1/videos/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async deleteVideo(id: string): Promise<void> {
    return this.request(`/v1/videos/${id}`, { method: 'DELETE' });
  }

  async likeVideo(id: string): Promise<{ liked: boolean; likeCount: number }> {
    return this.request(`/v1/videos/${id}/like`, { method: 'POST' });
  }

  async unlikeVideo(id: string): Promise<{ liked: boolean; likeCount: number }> {
    return this.request(`/v1/videos/${id}/like`, { method: 'DELETE' });
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
    return this.request(`/v1/search?${params.toString()}`);
  }

  async getSearchSuggestions(query: string): Promise<string[]> {
    return this.request(`/v1/search/suggestions?q=${encodeURIComponent(query)}`);
  }

  async getHomeFeed(page = 1, limit = 24, category?: string): Promise<{ items: Video[]; nextCursor?: string }> {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (category && category !== 'all') params.set('category', category);
    return this.request(`/v1/feed/home?${params.toString()}`);
  }

  async getShortsFeed(page = 1, limit = 10): Promise<{ items: Video[]; nextCursor?: string }> {
    return this.request(`/v1/feed/shorts?page=${page}&limit=${limit}`);
  }

  async getSubscriptionsFeed(page = 1, limit = 24): Promise<{ items: Video[]; total: number }> {
    return this.request(`/v1/feed/subscriptions?page=${page}&limit=${limit}`);
  }

  async sendFeedFeedback(videoId: string, action: 'not_interested' | 'dont_recommend_channel'): Promise<void> {
    return this.request('/v1/feed/feedback', {
      method: 'POST',
      body: JSON.stringify({ videoId, action }),
    });
  }

  // --- Comments ---
  async getVideoComments(videoId: string, page = 1, limit = 20, sort: 'top' | 'newest' = 'top'): Promise<{ items: Comment[]; total: number }> {
    return this.request(`/v1/videos/${videoId}/comments?page=${page}&limit=${limit}&sort=${sort}`);
  }

  async addComment(videoId: string, content: string, parentId?: string): Promise<Comment> {
    return this.request(`/v1/videos/${videoId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ content, parent_id: parentId }),
    });
  }

  async deleteComment(commentId: string): Promise<void> {
    return this.request(`/v1/comments/${commentId}`, { method: 'DELETE' });
  }

  async likeComment(commentId: string): Promise<{ liked: boolean; likeCount: number }> {
    return this.request(`/v1/comments/${commentId}/like`, { method: 'POST' });
  }

  async heartComment(commentId: string): Promise<{ isHearted: boolean }> {
    return this.request(`/v1/comments/${commentId}/heart`, { method: 'POST' });
  }

  async pinComment(commentId: string): Promise<{ isPinned: boolean }> {
    return this.request(`/v1/comments/${commentId}/pin`, { method: 'POST' });
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
    return this.request('/v1/live-streams', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getLiveStream(id: string): Promise<LiveStream> {
    return this.request(`/v1/live-streams/${id}`);
  }

  async startLiveStream(id: string): Promise<LiveStream> {
    return this.request(`/v1/live-streams/${id}/start`, { method: 'POST' });
  }

  async stopLiveStream(id: string): Promise<LiveStream> {
    return this.request(`/v1/live-streams/${id}/stop`, { method: 'POST' });
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
    return this.request('/v1/studio/dashboard');
  }

  async getStudioVideos(page = 1, limit = 20, status?: string): Promise<{ items: Video[]; total: number }> {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (status) params.set('status', status);
    return this.request(`/v1/studio/videos?${params.toString()}`);
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
    return this.request('/v1/ads/decision', {
      method: 'POST',
      body: JSON.stringify(data),
    });
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
    return this.request('/v1/ads/preview', {
      method: 'POST',
      body: JSON.stringify(data),
    });
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
    return this.request('/v1/partners/campaigns');
  }

  async createPartnerCampaign(data: Partial<AdCampaign>): Promise<AdCampaign> {
    return this.request('/v1/partners/campaigns', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getPartnerClaims(): Promise<CopyrightClaim[]> {
    return this.request('/v1/partners/rights/claims');
  }

  async createCopyrightClaim(data: { videoId: string; assetId: string; action: string }): Promise<CopyrightClaim> {
    return this.request('/v1/partners/rights/claims', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getPartnerSettlements(): Promise<Payout[]> {
    return this.request('/v1/partners/settlements');
  }

  // --- Admin ---
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
    return this.request(`/v1/admin/users?page=${page}&limit=${limit}`);
  }

  async updateAdminUserStatus(userId: string, status: 'active' | 'suspended' | 'banned', reason: string): Promise<User> {
    return this.request(`/v1/admin/users/${userId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status, reason }),
    });
  }

  async getAdminModerationQueue(): Promise<ModerationCase[]> {
    return this.request('/v1/admin/moderation');
  }

  async resolveModerationCase(caseId: string, action: 'approve' | 'reject' | 'delete', notes: string): Promise<ModerationCase> {
    return this.request(`/v1/admin/moderation/${caseId}/resolve`, {
      method: 'POST',
      body: JSON.stringify({ action, notes }),
    });
  }

  async getAdminAuditLogs(page = 1, limit = 50): Promise<{ items: Array<{ id: string; actorId: string; action: string; target: string; ip: string; timestamp: string }>; total: number }> {
    return this.request(`/v1/admin/audit-logs?page=${page}&limit=${limit}`);
  }
}

export function createApiClient(config: ApiClientConfig): ApiClient {
  return new ApiClient(config);
}

