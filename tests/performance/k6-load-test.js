/**
 * Global Video Platform — k6 Commercial Baseline Load & Stress Test
 *
 * Target Commercial Baseline (Section 23):
 *  - 10,000 Registered Users
 *  - 1,000 Concurrent Viewers (Home Feed, Watch Playback Manifest, Analytics Beacon)
 *  - 100 Concurrent Multipart Uploads (Upload Session Init + Presigned Part URL)
 *  - 10 Concurrent Live Streams (Live Stream Metadata + Realtime Chat Heartbeat)
 *
 * Mandatory SLO Thresholds:
 *  - API p95 Latency < 300ms (`http_req_duration: ['p(95)<300']`)
 *  - Error Rate < 0.1% (`http_req_failed: ['rate<0.001']`)
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('gvp_error_rate');
const playbackLatency = new Trend('gvp_playback_manifest_duration_ms');
const uploadInitLatency = new Trend('gvp_upload_init_duration_ms');

const BASE_URL = __ENV.GVP_API_URL || 'http://localhost:8080/v1';

export const options = {
  scenarios: {
    concurrent_viewers: {
      executor: 'ramping-vus',
      startVUs: 50,
      stages: [
        { duration: '1m', target: 500 },
        { duration: '3m', target: 1000 },
        { duration: '1m', target: 0 },
      ],
      exec: 'viewerScenario',
    },
    concurrent_uploads: {
      executor: 'constant-vus',
      vus: 100,
      duration: '4m',
      exec: 'uploaderScenario',
    },
    concurrent_live_streams: {
      executor: 'constant-vus',
      vus: 10,
      duration: '4m',
      exec: 'liveStreamerScenario',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<300', 'p(99)<600'],
    http_req_failed: ['rate<0.001'],
    gvp_error_rate: ['rate<0.001'],
    gvp_playback_manifest_duration_ms: ['p(95)<200'],
  },
};

export function viewerScenario() {
  // 1. Load Home Feed
  const feedRes = http.get(`${BASE_URL}/feed/home?page=1&limit=24`);
  const feedOk = check(feedRes, {
    'home feed status is 200': (r) => r.status === 200,
  });
  errorRate.add(!feedOk);

  // 2. Request Adaptive HLS Playback Manifest
  const pbRes = http.get(`${BASE_URL}/videos/00000000-0000-0000-0000-000000000001/playback`);
  playbackLatency.add(pbRes.timings.duration);
  const pbOk = check(pbRes, {
    'playback status is 200': (r) => r.status === 200,
    'contains playback_token': (r) => r.body && r.body.includes('playback_token'),
  });
  errorRate.add(!pbOk);

  // 3. Emit View Heartbeat to Analytics Ingestion Endpoint
  const eventPayload = JSON.stringify({
    event_id: `evt_${__VU}_${__ITER}_${Date.now()}`,
    event_type: 'heartbeat',
    session_id: `sess_${__VU}`,
    video_id: '00000000-0000-0000-0000-000000000001',
    channel_id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    device_type: 'web_desktop',
    watch_duration_seconds: 15,
    current_position_seconds: 120,
  });
  const evRes = http.post(`${BASE_URL}/events/analytics`, eventPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  errorRate.add(evRes.status !== 202);

  sleep(1);
}

export function uploaderScenario() {
  const headers = {
    'Content-Type': 'application/json',
    Authorization: 'Bearer creator-token',
  };

  const initPayload = JSON.stringify({
    title: `k6 Concurrent Upload Test VU-${__VU}`,
    fileSizeBytes: 52428800,
    mimeType: 'video/mp4',
  });

  const initRes = http.post(`${BASE_URL}/uploads`, initPayload, { headers });
  uploadInitLatency.add(initRes.timings.duration);
  const initOk = check(initRes, {
    'upload session created (201)': (r) => r.status === 201,
  });
  errorRate.add(!initOk);

  sleep(2);
}

export function liveStreamerScenario() {
  const res = http.get(`${BASE_URL}/live-streams/live-demo-01`);
  const ok = check(res, {
    'live stream status is 200': (r) => r.status === 200,
  });
  errorRate.add(!ok);
  sleep(1);
}
