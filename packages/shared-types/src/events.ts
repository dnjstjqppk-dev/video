export interface AnalyticsPlaybackEvent {
  event_id: string;
  event_type: 'impression' | 'click' | 'play_start' | 'progress' | 'valid_view' | 'complete' | 'like' | 'subscribe' | 'comment' | 'save' | 'dislike' | 'report';
  occurred_at: string;
  received_at?: string;
  user_id?: string;
  session_id: string;
  video_id: string;
  channel_id: string;
  device_type: 'web_desktop' | 'web_mobile' | 'ios' | 'android' | 'tv';
  client_version: string;
  playback_session_id: string;
  request_id?: string;
  watch_duration_seconds?: number;
  current_position_seconds?: number;
  video_duration_seconds?: number;
}

