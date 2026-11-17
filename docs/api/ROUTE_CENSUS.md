# API Route Census (generated, chi.Walk verified)

Source: `chi.Walk` dump of `services/api/cmd/server` (2026-11-17, 143 registrations). Generator: `python docs/generate_route_census.py` ??do not hand-edit.

Auth legend: **Public** serves visitors · **Auth** needs a session · **Role(...)** needs the role ·
**HOLE** is public but must be guarded (P0). `/v1/*/` in chi output is the users-router mount artifact
for real paths `/v1/me...`.

## ops (3)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | Public |  |
| `GET` | `/metrics` | Public |  |
| `GET` | `/ready` | Public |  |

## me (17)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/me` | Auth | chi renders `/v1/*/me` |
| `PATCH` | `/v1/me` | Auth | chi renders `/v1/*/me` |
| `GET` | `/v1/me/blocked-users` | Auth | chi renders `/v1/*/me/blocked-users` |
| `POST` | `/v1/me/blocked-users` | Auth | chi renders `/v1/*/me/blocked-users` |
| `DELETE` | `/v1/me/blocked-users/{targetId}` | Auth | chi renders `/v1/*/me/blocked-users/{targetId}` |
| `POST` | `/v1/me/delete-request` | Auth | chi renders `/v1/*/me/delete-request` |
| `DELETE` | `/v1/me/delete-request` | Auth | chi renders `/v1/*/me/delete-request` |
| `DELETE` | `/v1/me/history` | Auth | chi renders `/v1/*/me/history` |
| `GET` | `/v1/me/history` | Auth | chi renders `/v1/*/me/history` |
| `GET` | `/v1/me/mfa` | Auth | chi renders `/v1/*/me/mfa` |
| `POST` | `/v1/me/mfa/disable` | Auth | chi renders `/v1/*/me/mfa/disable` |
| `POST` | `/v1/me/mfa/setup` | Auth | chi renders `/v1/*/me/mfa/setup` |
| `POST` | `/v1/me/mfa/verify` | Auth | chi renders `/v1/*/me/mfa/verify` |
| `GET` | `/v1/me/sessions` | Auth | chi renders `/v1/*/me/sessions` |
| `DELETE` | `/v1/me/sessions/{id}` | Auth | chi renders `/v1/*/me/sessions/{id}` |
| `POST` | `/v1/me/sessions/{id}/revoke` | Auth | chi renders `/v1/*/me/sessions/{id}/revoke` |
| `GET` | `/v1/me/strikes` | Auth | chi renders `/v1/*/me/strikes` |

## channels (6)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `POST` | `/v1/channels/` | Auth | trailing-slash route |
| `GET` | `/v1/channels/{id}` | Public |  |
| `PATCH` | `/v1/channels/{id}` | Auth |  |
| `POST` | `/v1/channels/{id}/subscribe` | Auth |  |
| `DELETE` | `/v1/channels/{id}/subscribe` | Auth |  |
| `GET` | `/v1/channels/{id}/videos` | Public |  |

## videos (11)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/videos/{id}/comments` | Public |  |
| `POST` | `/v1/videos/{id}/comments` | Auth |  |
| `DELETE` | `/v1/videos/{id}` | Auth |  |
| `GET` | `/v1/videos/{id}` | Public |  |
| `PATCH` | `/v1/videos/{id}` | Auth |  |
| `POST` | `/v1/videos/{id}/captions` | Auth |  |
| `GET` | `/v1/videos/{id}/captions` | Auth |  |
| `DELETE` | `/v1/videos/{id}/captions/{capId}` | Auth |  |
| `DELETE` | `/v1/videos/{id}/like` | Auth |  |
| `POST` | `/v1/videos/{id}/like` | Auth |  |
| `GET` | `/v1/videos/{id}/playback` | Public |  |

## comments (4)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `DELETE` | `/v1/comments/{id}` | Auth |  |
| `POST` | `/v1/comments/{id}/heart` | Auth |  |
| `POST` | `/v1/comments/{id}/like` | Auth |  |
| `POST` | `/v1/comments/{id}/pin` | Auth |  |

## uploads (5)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `POST` | `/v1/uploads/` | Auth | trailing-slash route |
| `PUT` | `/v1/uploads/{id}/chunks/{partNumber}` | Auth |  |
| `POST` | `/v1/uploads/{id}/complete` | Auth |  |
| `POST` | `/v1/uploads/{id}/parts` | Auth |  |
| `GET` | `/v1/uploads/{id}/status` | Auth |  |

## feed (4)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `POST` | `/v1/feed/feedback` | Public |  |
| `GET` | `/v1/feed/home` | Public |  |
| `GET` | `/v1/feed/shorts` | Public |  |
| `GET` | `/v1/feed/subscriptions` | Public |  |

## search (2)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/search` | Public |  |
| `GET` | `/v1/search/suggestions` | Public |  |

## live (15)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/live-streams/` | Public | trailing-slash route |
| `POST` | `/v1/live-streams/` | Auth | trailing-slash route |
| `GET` | `/v1/live-streams/{id}` | Public |  |
| `POST` | `/v1/live-streams/{id}/chat` | Auth |  |
| `GET` | `/v1/live-streams/{id}/chat` | Public |  |
| `GET` | `/v1/live-streams/{id}/health` | Public |  |
| `POST` | `/v1/live-streams/{id}/polls` | Auth |  |
| `GET` | `/v1/live-streams/{id}/polls/active` | Public |  |
| `POST` | `/v1/live-streams/{id}/polls/vote` | Auth |  |
| `POST` | `/v1/live-streams/{id}/rotate-key` | Auth |  |
| `POST` | `/v1/live-streams/{id}/start` | Auth |  |
| `POST` | `/v1/live-streams/{id}/stop` | Auth |  |
| `POST` | `/v1/live-streams/{id}/superchat` | Auth |  |
| `GET` | `/v1/live-streams/{id}/superchats` | Public |  |
| `GET` | `/v1/live-streams/{id}/ws` | Public |  |

## studio (9)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/studio/earnings` | Auth |  |
| `GET` | `/v1/studio/settlements` | Auth |  |
| `GET` | `/v1/studio/analytics` | Auth |  |
| `GET` | `/v1/studio/claims` | Auth |  |
| `GET` | `/v1/studio/comments` | Auth |  |
| `GET` | `/v1/studio/dashboard` | Auth |  |
| `GET` | `/v1/studio/videos` | Auth |  |
| `POST` | `/v1/studio/videos/{id}/edit-jobs` | Auth |  |
| `GET` | `/v1/studio/videos/{id}/edit-jobs/{jobId}` | Auth |  |

## partners (15)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/partners/campaigns` | Role(advertiser, partner_manager, platform_admin) |  |
| `POST` | `/v1/partners/campaigns` | Role(advertiser, partner_manager, platform_admin) |  |
| `POST` | `/v1/partners/rights/claims` | Auth |  |
| `GET` | `/v1/partners/rights/claims` | Auth |  |
| `POST` | `/v1/partners/rights/claims/{id}/dispute` | Auth |  |
| `POST` | `/v1/partners/rights/claims/{id}/resolve` | Role(super_admin, platform_admin, moderator) |  |
| `POST` | `/v1/partners/rights/dmca-counter-notice` | Auth |  |
| `POST` | `/v1/partners/rights/dmca-takedown` | Auth |  |
| `GET` | `/v1/partners/settlements` | Auth |  |
| `GET` | `/v1/partners/settlements/earnings-summary` | Auth |  |
| `POST` | `/v1/partners/settlements/payout` | Auth |  |
| `GET` | `/v1/partners/channels` | Role(partner_manager, advertiser, rights_holder, rights_owner, partner_finance, mcn_manager, partner_admin, platform_admin) |  |
| `GET` | `/v1/partners/contracts` | Role(partner_manager, advertiser, rights_holder, rights_owner, partner_finance, mcn_manager, partner_admin, platform_admin) |  |
| `PATCH` | `/v1/partners/contracts/{id}` | Role(partner_manager, advertiser, rights_holder, rights_owner, partner_finance, mcn_manager, partner_admin, platform_admin) |  |
| `GET` | `/v1/partners/me` | Role(partner_manager, advertiser, rights_holder, rights_owner, partner_finance, mcn_manager, partner_admin, platform_admin) |  |

## ads (15)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/ads/admin/creatives` | Role(super_admin, platform_admin) |  |
| `POST` | `/v1/ads/admin/creatives/{id}/approve` | Role(super_admin, platform_admin) |  |
| `POST` | `/v1/ads/admin/creatives/{id}/reject` | Role(super_admin, platform_admin) |  |
| `GET` | `/v1/ads/admin/diagnostics` | Role(super_admin, platform_admin) |  |
| `GET` | `/v1/ads/advertiser/reports` | Role(advertiser, partner_manager, platform_admin) |  |
| `GET` | `/v1/ads/campaigns` | Role(advertiser, partner_manager, platform_admin) |  |
| `POST` | `/v1/ads/campaigns` | Role(advertiser, partner_manager, platform_admin) |  |
| `GET` | `/v1/ads/campaigns/{id}` | Role(advertiser, partner_manager, platform_admin) |  |
| `POST` | `/v1/ads/campaigns/{id}/creatives` | Role(advertiser, partner_manager, platform_admin) |  |
| `PATCH` | `/v1/ads/campaigns/{id}/status` | Role(advertiser, partner_manager, platform_admin) |  |
| `GET` | `/v1/ads/creatives/{id}/asset` | Public |  |
| `POST` | `/v1/ads/decision` | Public |  |
| `POST` | `/v1/ads/events` | Public |  |
| `POST` | `/v1/ads/preview` | Role(advertiser, partner_manager, platform_admin) |  |
| `GET` | `/v1/ads/reports` | Role(advertiser, partner_manager, platform_admin) |  |

## advertiser (9)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/advertiser/campaigns` | Role(advertiser, partner_manager, platform_admin) |  |
| `POST` | `/v1/advertiser/campaigns` | Role(advertiser, partner_manager, platform_admin) |  |
| `GET` | `/v1/advertiser/campaigns/{id}` | Role(advertiser, partner_manager, platform_admin) |  |
| `POST` | `/v1/advertiser/campaigns/{id}/creatives` | Role(advertiser, partner_manager, platform_admin) |  |
| `PATCH` | `/v1/advertiser/campaigns/{id}/status` | Role(advertiser, partner_manager, platform_admin) |  |
| `POST` | `/v1/advertiser/decision` | Public |  |
| `POST` | `/v1/advertiser/events` | Public |  |
| `POST` | `/v1/advertiser/preview` | Role(advertiser, partner_manager, platform_admin) |  |
| `GET` | `/v1/advertiser/reports` | Role(advertiser, partner_manager, platform_admin) |  |

## admin (13)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/admin/ads/delivery-diagnostics` | Role(super_admin, platform_admin) |  |
| `POST` | `/v1/admin/finance/batches/approve` | Role(super_admin, platform_admin, finance_admin) |  |
| `GET` | `/v1/admin/finance/settlements` | Role(super_admin, platform_admin, finance_admin) |  |
| `GET` | `/v1/admin/finance/summary` | Role(super_admin, platform_admin, finance_admin) |  |
| `GET` | `/v1/admin/moderation` | Role(super_admin, platform_admin, moderator, support_agent) |  |
| `POST` | `/v1/admin/moderation/{id}/resolve` | Role(super_admin, platform_admin, moderator, support_agent) |  |
| `GET` | `/v1/admin/policies/` | Role(super_admin, platform_admin (+support read on GET)) | trailing-slash route |
| `PATCH` | `/v1/admin/policies/` | Role(super_admin, platform_admin (+support read on GET)) | trailing-slash route |
| `PATCH` | `/v1/admin/policies/flags/{key}` | Role(super_admin, platform_admin (+support read on GET)) |  |
| `GET` | `/v1/admin/audit-logs` | Role(super_admin, platform_admin, finance_admin, moderator, support_agent) |  |
| `GET` | `/v1/admin/dashboard` | Role(super_admin, platform_admin, finance_admin, moderator, support_agent) |  |
| `GET` | `/v1/admin/users` | Role(super_admin, platform_admin (+support read on GET /users)) |  |
| `PATCH` | `/v1/admin/users/{id}/status` | Role(super_admin, platform_admin (+support read on GET /users)) |  |

## payments (4)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `POST` | `/v1/payments/checkout` | Auth |  |
| `POST` | `/v1/payments/confirm` | Auth |  |
| `GET` | `/v1/payments/transactions` | Auth |  |
| `POST` | `/v1/payments/webhooks/{provider}` | Public |  |

## playlists (4)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `POST` | `/v1/playlists/` | Auth | trailing-slash route |
| `GET` | `/v1/playlists/{id}` | Public |  |
| `PATCH` | `/v1/playlists/{id}` | Auth |  |
| `POST` | `/v1/playlists/{id}/items` | Auth |  |

## subscriptions (3)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/subscriptions/` | Auth | trailing-slash route |
| `POST` | `/v1/subscriptions/{channelId}` | Auth |  |
| `DELETE` | `/v1/subscriptions/{channelId}` | Auth |  |

## moderation (1)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `POST` | `/v1/moderation/reports` | Auth |  |

## events (1)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `POST` | `/v1/events/analytics` | Public |  |

## notifications (2)

| Method | Path | Auth | Notes |
| :--- | :--- | :--- | :--- |
| `GET` | `/v1/notifications/` | Auth | trailing-slash route |
| `POST` | `/v1/notifications/{id}/read` | Auth |  |
