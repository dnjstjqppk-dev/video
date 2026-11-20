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

- Adaptive player: overlay controls with auto-hide, segmented chapter seekbar, settings menu (quality from real ABR rungs, speed `0.25x`-`2.0x`, captions, autoplay, stable volume, ambient, sleep timer, stats), captions, chapter
  bar, mini player (PiP, `I`), theater mode (`T`), fullscreen and shortcuts (`Space`, `K`,
  `J`, `L`, `M`, `F`, `C`, `0`-`9`, `,`/`.` frame step, `Shift+S` stats HUD).
- Below the player: subscribe with notification options, like/dislike, share,
  save and report actions, AI badge for AI-made videos, series chip for episodes,
  expandable description with timestamp chapters, and threaded comments with sort
  and timestamp seeking.
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
- Content (`/content`): type tabs (동영상/일반/숏폼/라이브) plus processing-status
  and view-count filters; bulk delete and add-to-playlist; per-row visibility
  selector; quick actions for edit, analytics, comments and watch-on-web; AI
  flag toggle + review-submit + status badge (미제출/심사중/선정/반려).
- Upload (`/upload`): drag-and-drop multipart upload (10 MB chunks, pause/
  resume, per-chunk retry) with progress, then the 4-step wizard (details,
  elements, checks, visibility/schedule) with custom thumbnails, AI metadata
  (AI-made flag + tool) and a DRM checkbox.
- Series (`/series`): season/episode attach/detach and completed status for
  creator-run binge series.
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
- AI reviews (`/ai-reviews`): Originals review queue (approve/reject with memo).
- Sensitive actions always require the reviewer's own 6-digit TOTP code and
  are recorded in the audit log.

## 4. AI Originals hall (web console)

- Shelf (`/ai`): always-dark cinema hall with a featured billboard (Play / More
  Info) and rails for binge series, Originals picks and AI Shorts.
- Series (`/series/:id`): dark hero with binge-from-E1, season-grouped episode
  list with AI badges.
- Watch: episode series panel (current-episode mark, expand), next-episode
  autoplay preference, AI badge with tool name.
- Creator flow: Studio upload/content AI flag → submit for review → admin
  approves in `/ai-reviews` → Originals shelf exposure. Verified end to end
  on 2026-11-20 (real upload → badge → review → binge → autoplay).
