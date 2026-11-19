# vravra User Guide — Viewer, Creator and Admin Manuals

## 1. Viewer (web console, port 3000)

### Home feed (`/`)

- Filter with the sticky category rail: All, Music, Gaming, News, Live, Tech,
  Finance, Travel, Recently uploaded, Watched. Use the side arrows to scroll.
- Each card shows duration, watch progress, views and upload age. Hover for a
  muted preview (when available). The `⋮` menu offers Save to Watch Later,
  Share, Not interested, Don't recommend channel, and Report.
- Live Now and Shorts shelves link to full live and shorts surfaces.

### Watch page (`/watch/:id`)

- Adaptive player: quality selector, speed (`0.25x`-`2.0x`), captions, chapter
  bar, mini player (PiP), theater mode, fullscreen and shortcuts (`Space`, `K`,
  `J`, `L`, `M`, `F`, `C`, `0`-`9`).
- Below the player: subscribe with notification options, like/dislike, share,
  save, clip and a `•••` overflow menu, expandable description with timestamp
  chapters, and threaded comments with sort and timestamp seeking.
- The right rail lists related videos with filter chips.

### Search, shorts, live and channels

- Search (`/search`) supports type filters; the header box keeps recent terms
  (hover to delete one) and live suggestions with keyboard navigation.
- Shorts (`/shorts`) is a fullscreen vertical feed with like, comment, share
  and audio actions. Live (`/live/:id`) adds chat, slow mode, SuperChat and
  polls. Channel pages (`/channels/:id`) group home, videos, shorts, live,
  playlists, community and about tabs.

### Account and workspace

- The bell opens the notification center (All, Subscriptions, Comments,
  Studio tabs) with unread dots and mark-all-read.
- The avatar menu switches workspace (Viewer, Partner Studio, Admin Console),
  toggles dark/light theme and language, and signs out. The `+ Create` menu
  deep-links into Studio upload, shorts, live and community flows.

## 2. Creator (Studio console, port 3001)

- Dashboard (`/`): latest video performance, 28-day channel summary, recent
  comments with inline replies, and platform notices.
- Content (`/content`): type tabs (Videos, Standard, Shorts, Live) plus a
  processing-status filter; bulk visibility changes; per-row visibility
  selector; quick actions for edit, analytics, comments and watch-on-web.
- Upload (`/upload`): drag-and-drop multipart upload with progress, then the
  4-step wizard (details, elements, checks, visibility/schedule) with custom
  thumbnails.
- Analytics (`/analytics`): 7/28/90/365-day ranges, views and watch-time
  series charts, revenue trend, and CSV export.
- Comments (`/comments`): published/pending/unanswered/question filters with
  inline reply, heart, pin and hide/report actions.
- Earn (`/monetization`): partner eligibility progress, revenue sources,
  settlement statements and payout requests.
- Extras: timeline editor (`/editor/:id`), live room (`/live`), subtitles,
  copyright disputes and channel customization.

## 3. Admin (console, port 3003)

- Overview (`/`): concurrent viewers, bandwidth, uploads, revenue and pending
  reports with service health indicators.
- Moderation (`/moderation`): pending/resolved filter, split review cards with
  report detail on the left and notes + TOTP + one-click actions (approve,
  19+ restrict, takedown with evidence) on the right; video cases link to the
  source page.
- Users (`/users`): search, role and status views, sanctions (suspend/ban/
  reactivate) with mandatory reason + TOTP, and a recent audit trail.
- Finance (`/finance`): double-entry batch approval with balance guard and
  TOTP, settlement statements with CSV export, and a cost simulator.
- Policies (`/policies`): policy versions, kill-switch flags and audit logs.
- Ads (`/ads`): delivery diagnostics, campaign review and creative approval.
- Sensitive actions always require the reviewer's own 6-digit TOTP code and
  are recorded in the immutable audit log.
