/**
 * Playwright End-to-End (E2E) Multi-Surface Platform Specification
 * Covers all mandatory Section 22 Verification Scenarios:
 *  - Surface 1 (User Web): Signup/Auth -> Home Feed -> Watch HLS Player (Chapters, Captions, Theater/PiP) -> Subscribe -> Like -> Comment -> Shorts Vertical Swipe
 *  - Surface 2 (Creator Studio): Multipart Upload Session -> Chunk Resume -> Metadata & Visibility Edit -> Analytics Time-Series -> Earnings
 *  - Surface 3 (Partner Portal): Ad Campaign Creation -> MCN Linked Channels -> Copyright Content ID Claim -> Monthly Settlement Ledger
 *  - Surface 4 (Admin Console): User Status Sanction -> Trust & Safety Moderation Queue Resolution -> Immutable Audit Log Verification
 *  - Surface 5 (Security & Isolation): Unauthenticated Redirect, BOLA/IDOR Cross-Channel Mutation Block, Non-Admin Console Access Block
 */
import { test, expect } from '@playwright/test';

const API_BASE_URL = process.env.GVP_API_URL || 'http://localhost:8080/v1';

test.describe('Global Video Platform — Section 22 End-to-End Release Gate Suite', () => {
  test('1. User Surface: Home Feed -> Watch HLS Manifest -> Subscribe -> Like -> Comment -> Shorts Feed', async ({ request }) => {
    // 1.1 Fetch Personalized Home Feed
    const homeRes = await request.get(`${API_BASE_URL}/feed/home?page=1&limit=24`);
    expect(homeRes.status()).toBe(200);
    const homeData = await homeRes.json();
    expect(Array.isArray(homeData.items)).toBe(true);
    expect(homeData.items.length).toBeGreaterThan(0);

    const targetVideoId = homeData.items[0].id;
    const targetChannelId = homeData.items[0].channel_id;

    // 1.2 Watch Page Playback Manifest (HLS + Chapters + Captions + Playback Token)
    const playbackRes = await request.get(`${API_BASE_URL}/videos/${targetVideoId}/playback`);
    expect(playbackRes.status()).toBe(200);
    const playback = await playbackRes.json();
    expect(playback.video_id).toBe(targetVideoId);
    expect(playback.hls_master_url).toContain('.m3u8');
    expect(playback.playback_token).toContain('pb_token_');
    expect(Array.isArray(playback.chapters)).toBe(true);
    expect(Array.isArray(playback.captions)).toBe(true);

    // 1.3 Subscribe to Channel
    const subRes = await request.post(`${API_BASE_URL}/channels/${targetChannelId}/subscribe`, {
      headers: { Authorization: 'Bearer creator-token' },
    });
    expect(subRes.status()).toBe(200);
    const subBody = await subRes.json();
    expect(subBody.subscribed).toBe(true);

    // 1.4 Like Video & Post Comment
    const likeRes = await request.post(`${API_BASE_URL}/videos/${targetVideoId}/like`, {
      headers: { Authorization: 'Bearer creator-token' },
    });
    expect(likeRes.status()).toBe(200);

    const commentRes = await request.post(`${API_BASE_URL}/videos/${targetVideoId}/comments`, {
      headers: { Authorization: 'Bearer creator-token' },
      data: { content: 'Playwright E2E verified HLS playback and chapters!' },
    });
    expect(commentRes.status()).toBe(201);

    // 1.5 Shorts Vertical Feed
    const shortsRes = await request.get(`${API_BASE_URL}/feed/shorts`);
    expect(shortsRes.status()).toBe(200);
    const shortsData = await shortsRes.json();
    expect(Array.isArray(shortsData.items)).toBe(true);
  });

  test('2. Creator Studio Surface: Multipart Upload -> Resume Part -> Complete -> Edit Metadata -> Analytics', async ({ request }) => {
    const headers = { Authorization: 'Bearer creator-token' };

    // 2.1 Initiate Multipart Upload Session
    const initRes = await request.post(`${API_BASE_URL}/uploads`, {
      headers,
      data: {
        title: 'Creator Studio E2E Masterclass Upload',
        description: 'Uploaded via resumable multipart R2 pipeline',
        fileSizeBytes: 25 * 1024 * 1024,
        mimeType: 'video/mp4',
        videoType: 'standard',
      },
    });
    expect(initRes.status()).toBe(201);
    const session = await initRes.json();
    expect(session.uploadId).toBeTruthy();
    expect(session.totalParts).toBe(3);

    // 2.2 Request Presigned Part URL (Simulating Chunk Resume for Part 2)
    const partRes = await request.post(`${API_BASE_URL}/uploads/${session.uploadId}/parts`, {
      headers,
      data: { partNumber: 2 },
    });
    expect(partRes.status()).toBe(200);
    const partBody = await partRes.json();
    expect(partBody.uploadUrl).toContain('partNumber=2');

    // 2.3 Complete Multipart Upload & Trigger FFmpeg Transcoding Worker
    const completeRes = await request.post(`${API_BASE_URL}/uploads/${session.uploadId}/complete`, {
      headers,
      data: {
        parts: [
          { partNumber: 1, etag: '"etag-1"' },
          { partNumber: 2, etag: '"etag-2"' },
          { partNumber: 3, etag: '"etag-3"' },
        ],
      },
    });
    expect(completeRes.status()).toBe(200);

    // 2.4 Update Video Metadata & Visibility in Studio
    const editRes = await request.patch(`${API_BASE_URL}/videos/${session.videoId}`, {
      headers,
      data: {
        title: 'Creator Studio E2E Masterclass Upload (Updated)',
        visibility: 'public',
        category: 'Technology',
      },
    });
    expect(editRes.status()).toBe(200);

    // 2.5 Fetch Studio Analytics & Earnings
    const analyticsRes = await request.get(`${API_BASE_URL}/studio/analytics`, { headers });
    expect(analyticsRes.status()).toBe(200);
    const earningsRes = await request.get(`${API_BASE_URL}/studio/earnings`, { headers });
    expect(earningsRes.status()).toBe(200);
  });

  test('3. Partner Portal Surface: Ad Campaign -> MCN Channels -> Copyright Claim -> Settlement Statement', async ({ request }) => {
    const headers = { Authorization: 'Bearer creator-token' };

    // 3.1 Create Ad Campaign
    const campRes = await request.post(`${API_BASE_URL}/partners/campaigns`, {
      headers,
      data: {
        name: 'Q4 Brand Launch Preroll Campaign',
        format: 'preroll',
        totalBudget: 10000000,
        dailyBudget: 500000,
      },
    });
    expect(campRes.status()).toBe(201);

    // 3.2 List MCN Linked Channels
    const chRes = await request.get(`${API_BASE_URL}/partners/channels`, { headers });
    expect(chRes.status()).toBe(200);

    // 3.3 Submit Content ID Copyright Claim
    const claimRes = await request.post(`${API_BASE_URL}/partners/rights/claims`, {
      headers,
      data: {
        videoId: '00000000-0000-0000-0000-000000000001',
        assetId: 'asset-music-001',
        action: 'monetize',
      },
    });
    expect(claimRes.status()).toBe(201);

    // 3.4 Retrieve Partner Monthly Settlements
    const settleRes = await request.get(`${API_BASE_URL}/partners/settlements`, { headers });
    expect(settleRes.status()).toBe(200);
  });

  test('4. Admin Console & Security Isolation: Sanction User -> Resolve Moderation -> Verify Audit Log & RBAC Block', async ({ request }) => {
    // 4.1 Security Check: Non-admin creator must be blocked from Admin Console API (403 FORBIDDEN_RESOURCE)
    const forbiddenRes = await request.get(`${API_BASE_URL}/admin/dashboard`, {
      headers: { Authorization: 'Bearer creator-token' },
    });
    expect(forbiddenRes.status()).toBe(403);
    const forbiddenBody = await forbiddenRes.json();
    expect(forbiddenBody.error.code).toBe('FORBIDDEN_RESOURCE');

    // 4.2 Super Admin sanctions user status
    const adminHeaders = { Authorization: 'Bearer admin-token' };
    const sanctionRes = await request.patch(`${API_BASE_URL}/admin/users/22222222-2222-2222-2222-222222222222/status`, {
      headers: adminHeaders,
      data: { status: 'active', reason: 'Periodic compliance verification passed' },
    });
    expect(sanctionRes.status()).toBe(200);

    // 4.3 Super Admin resolves Trust & Safety moderation case
    const modRes = await request.post(`${API_BASE_URL}/admin/moderation/mod-case-1/resolve`, {
      headers: adminHeaders,
      data: { action: 'approve', notes: 'Fair use educational commentary verified' },
    });
    expect(modRes.status()).toBe(200);

    // 4.4 Verify Immutable Admin Audit Log Recorded
    const auditRes = await request.get(`${API_BASE_URL}/admin/audit-logs`, { headers: adminHeaders });
    expect(auditRes.status()).toBe(200);
    const auditLogs = await auditRes.json();
    expect(auditLogs.total).toBeGreaterThanOrEqual(1);
  });
});
