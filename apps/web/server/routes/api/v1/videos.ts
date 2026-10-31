export default defineEventHandler((event) => {
  setResponseHeaders(event, {
    'Content-Type': 'application/json; charset=utf-8',
    'Cache-Control': 'public, max-age=60',
    'Access-Control-Allow-Origin': '*'
  })

  const videos = [
    {
      id: 'vid_arch_2026',
      channelId: 'ch_tech_kr',
      channel: {
        id: 'ch_tech_kr',
        ownerId: 'usr_1',
        handle: '@gvptech',
        name: 'GVP Tech Engineering',
        avatarUrl: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=120&q=80',
        subscriberCount: 284000,
        videoCount: 142,
        isVerified: true,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z'
      },
      title: 'Building a 10M Concurrent Viewer Global Video Platform with Go 1.24, Nuxt 4 & ClickHouse',
      description: 'Deep-dive into HLS ABR transcoding, NATS JetStream event pipelines, and OpenSearch hybrid ranking.',
      videoType: 'standard',
      durationSeconds: 1845,
      thumbnailUrl: 'https://images.unsplash.com/photo-1518770660439-4636190af475?w=800&q=80',
      visibility: 'public',
      status: 'ready',
      viewCount: 412900,
      likeCount: 32400,
      dislikeCount: 120,
      commentCount: 1842,
      tags: ['golang', 'nuxt4', 'system-design'],
      category: 'tech',
      isAgeRestricted: false,
      sources: [],
      chapters: [],
      captions: [],
      publishedAt: '2026-10-20T09:00:00Z',
      createdAt: '2026-10-20T08:00:00Z',
      updatedAt: '2026-10-20T09:00:00Z'
    },
    {
      id: 'vid_seoul_4k',
      channelId: 'ch_seoul_4k',
      channel: {
        id: 'ch_seoul_4k',
        ownerId: 'usr_2',
        handle: '@seoul4k',
        name: 'Seoul 4K Cinema',
        avatarUrl: 'https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=120&q=80',
        subscriberCount: 890000,
        videoCount: 310,
        isVerified: true,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z'
      },
      title: '[4K 60fps HDR] Cyberpunk Rainy Night Walk in Gangnam & Hongdae, Seoul 2026',
      description: 'Binaural 3D audio night walk in Seoul captured in 2160p HDR10.',
      videoType: 'standard',
      durationSeconds: 3620,
      thumbnailUrl: 'https://images.unsplash.com/photo-1517154421773-0529f29ea451?w=800&q=80',
      visibility: 'public',
      status: 'ready',
      viewCount: 1294000,
      likeCount: 95100,
      dislikeCount: 310,
      commentCount: 4510,
      tags: ['seoul', '4k', 'hdr'],
      category: 'cinema',
      isAgeRestricted: false,
      sources: [],
      chapters: [],
      captions: [],
      publishedAt: '2026-10-18T14:00:00Z',
      createdAt: '2026-10-18T12:00:00Z',
      updatedAt: '2026-10-18T14:00:00Z'
    },
    {
      id: 'vid_esports_finals',
      channelId: 'ch_esports_global',
      channel: {
        id: 'ch_esports_global',
        ownerId: 'usr_3',
        handle: '@esports',
        name: 'Global Esports Arena',
        avatarUrl: 'https://images.unsplash.com/photo-1566492031773-4f4e44671857?w=120&q=80',
        subscriberCount: 2450000,
        videoCount: 980,
        isVerified: true,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z'
      },
      title: 'T1 vs Gen.G — Grand Finals Game 5 Full Highlights (Korean & English Dual Cast)',
      description: 'Unforgettable final teamfight at Baron pit in 1440p 60fps.',
      videoType: 'standard',
      durationSeconds: 1490,
      thumbnailUrl: 'https://images.unsplash.com/photo-1542751371-adc38448a05e?w=800&q=80',
      visibility: 'public',
      status: 'ready',
      viewCount: 3480000,
      likeCount: 215000,
      dislikeCount: 1900,
      commentCount: 19400,
      tags: ['esports', 'gaming', 'finals'],
      category: 'gaming',
      isAgeRestricted: false,
      sources: [],
      chapters: [],
      captions: [],
      publishedAt: '2026-10-22T18:30:00Z',
      createdAt: '2026-10-22T17:00:00Z',
      updatedAt: '2026-10-22T18:30:00Z'
    }
  ]

  return {
    items: videos,
    total: videos.length,
    page: 1,
    limit: 24,
    hasMore: false
  }
})
