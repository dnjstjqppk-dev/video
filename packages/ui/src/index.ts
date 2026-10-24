export interface VideoCardProps {
  id: string;
  title: string;
  thumbnailUrl: string;
  durationSeconds: number;
  viewCount: number;
  publishedAt: string;
  channelName: string;
  channelHandle: string;
  channelAvatarUrl?: string;
  isLive?: boolean;
  isShorts?: boolean;
}

export interface StatusBadgeProps {
  status: 'READY' | 'PROCESSING' | 'CREATED' | 'FAILED' | 'LIVE' | 'PAID' | 'PENDING' | string;
  size?: 'sm' | 'md' | 'lg';
}

export interface MetricCardProps {
  label: string;
  value: string | number;
  deltaPercent?: number;
  trend?: 'up' | 'down' | 'neutral';
  currency?: string;
}

export interface ShakaPlayerConfig {
  abrEnabled: boolean;
  defaultBandwidthEstimate: number;
  switchIntervalSeconds: number;
  bufferingGoalSeconds: number;
  rebufferingGoalSeconds: number;
  lowLatencyMode: boolean;
  preferNativeHlsOnSafari: boolean;
}

export const DEFAULT_SHAKA_PLAYER_CONFIG: ShakaPlayerConfig = {
  abrEnabled: true,
  defaultBandwidthEstimate: 5_000_000,
  switchIntervalSeconds: 2,
  bufferingGoalSeconds: 20,
  rebufferingGoalSeconds: 2,
  lowLatencyMode: false,
  preferNativeHlsOnSafari: true,
};

export function formatDurationBadge(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) return '0:00';
  const total = Math.floor(seconds);
  const hrs = Math.floor(total / 3600);
  const mins = Math.floor((total % 3600) / 60);
  const secs = total % 60;
  const paddedSecs = secs < 10 ? `0${secs}` : `${secs}`;
  if (hrs > 0) {
    const paddedMins = mins < 10 ? `0${mins}` : `${mins}`;
    return `${hrs}:${paddedMins}:${paddedSecs}`;
  }
  return `${mins}:${paddedSecs}`;
}

export function resolveStatusBadgeTone(status: string): 'success' | 'warning' | 'danger' | 'info' | 'neutral' {
  switch (status.toUpperCase()) {
    case 'READY':
    case 'PAID':
    case 'ACTIVE':
    case 'HEALTHY':
      return 'success';
    case 'PROCESSING':
    case 'VALIDATING':
    case 'QUEUED':
    case 'PENDING':
      return 'warning';
    case 'FAILED':
    case 'BANNED':
    case 'REJECTED':
      return 'danger';
    case 'LIVE':
      return 'info';
    default:
      return 'neutral';
  }
}
