import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/network/api_client.dart';

class HomeVideoItem {
  final String id;
  final String title;
  final String channelName;
  final String category;
  final String viewsText;
  final String publishedText;
  final String durationLabel;
  final String? thumbnailUrl;
  final bool isLive;
  final Color accentColor;

  const HomeVideoItem({
    required this.id,
    required this.title,
    required this.channelName,
    required this.category,
    required this.viewsText,
    required this.publishedText,
    required this.durationLabel,
    this.thumbnailUrl,
    this.isLive = false,
    required this.accentColor,
  });

  factory HomeVideoItem.fromApi(Map<String, dynamic> json) {
    final viewCount = (json['view_count'] as num?)?.toInt() ?? 0;
    final durationSeconds = (json['duration_seconds'] as num?)?.toInt() ?? 0;
    final videoType = (json['video_type'] as String?) ?? 'standard';
    final publishedRaw =
        (json['published_at'] ?? json['created_at']) as String?;
    return HomeVideoItem(
      id: (json['id'] as String?) ?? '',
      title: (json['title'] as String?) ?? '(untitled)',
      channelName: (json['channel_name'] as String?) ?? 'Unknown channel',
      category: (json['category'] as String?) ?? 'General',
      viewsText: _formatViews(viewCount),
      publishedText: _formatPublished(publishedRaw),
      durationLabel: videoType == 'live' ? 'LIVE' : _formatDuration(durationSeconds),
      thumbnailUrl: json['thumbnail_url'] as String?,
      isLive: videoType == 'live',
      accentColor: const Color(0xFFE11D48),
    );
  }

  static String _formatViews(int count) {
    if (count >= 1000000) return '${(count / 1000000).toStringAsFixed(1)}M views';
    if (count >= 1000) return '${(count / 1000).toStringAsFixed(1)}K views';
    return '$count views';
  }

  static String _formatPublished(String? iso) {
    if (iso == null || iso.isEmpty) return '';
    final date = DateTime.tryParse(iso);
    if (date == null) return '';
    final diff = DateTime.now().difference(date);
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    if (diff.inDays < 30) return '${diff.inDays}d ago';
    return '${date.year}.${date.month.toString().padLeft(2, '0')}.${date.day.toString().padLeft(2, '0')}';
  }

  static String _formatDuration(int seconds) {
    final m = seconds ~/ 60;
    final s = seconds % 60;
    return '$m:${s.toString().padLeft(2, '0')}';
  }
}

class HomeFeedState {
  final String selectedCategory;
  final List<HomeVideoItem> items;
  final bool isLoading;
  final bool isLoadingMore;
  final int page;
  final bool hasMore;
  final String? errorMessage;

  const HomeFeedState({
    this.selectedCategory = 'All',
    this.items = const [],
    this.isLoading = true,
    this.isLoadingMore = false,
    this.page = 1,
    this.hasMore = true,
    this.errorMessage,
  });

  List<HomeVideoItem> get filteredItems {
    if (selectedCategory == 'All') return items;
    if (selectedCategory == 'Live') {
      return items.where((item) => item.isLive).toList();
    }
    return items.where((item) => item.category == selectedCategory).toList();
  }

  HomeFeedState copyWith({
    String? selectedCategory,
    List<HomeVideoItem>? items,
    bool? isLoading,
    bool? isLoadingMore,
    int? page,
    bool? hasMore,
    String? Function()? errorMessage,
  }) {
    return HomeFeedState(
      selectedCategory: selectedCategory ?? this.selectedCategory,
      items: items ?? this.items,
      isLoading: isLoading ?? this.isLoading,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
      page: page ?? this.page,
      hasMore: hasMore ?? this.hasMore,
      errorMessage: errorMessage != null ? errorMessage() : this.errorMessage,
    );
  }
}

final homeFeedProvider =
    StateNotifierProvider<HomeFeedNotifier, HomeFeedState>((ref) {
  final api = ref.watch(apiClientProvider);
  return HomeFeedNotifier(api)..loadInitial();
});

class HomeFeedNotifier extends StateNotifier<HomeFeedState> {
  final ApiClient _api;

  HomeFeedNotifier(this._api) : super(const HomeFeedState());

  static const List<String> categories = [
    'All',
    'Live',
    'Engineering',
    'Gaming',
    'Music',
    'Cinema 4K',
    'AI & Cloud',
  ];

  static const int _pageSize = 20;

  Future<void> loadInitial() async {
    state = state.copyWith(isLoading: true, errorMessage: () => null);
    try {
      final res = await _api.get<Map<String, dynamic>>(
        '/feed/home',
        queryParameters: {'page': 1, 'limit': _pageSize},
      );
      final data = res.data ?? {};
      final items = _parseItems(data);
      final total = (data['total'] as num?)?.toInt() ?? items.length;
      state = state.copyWith(
        items: items,
        page: 1,
        hasMore: items.length < total,
        isLoading: false,
      );
    } on ApiErrorException catch (e) {
      state = state.copyWith(isLoading: false, errorMessage: () => e.message);
    }
  }

  void selectCategory(String category) {
    state = state.copyWith(selectedCategory: category);
  }

  Future<void> refreshFeed() async {
    state = state.copyWith(errorMessage: () => null);
    try {
      final res = await _api.get<Map<String, dynamic>>(
        '/feed/home',
        queryParameters: {'page': 1, 'limit': _pageSize},
      );
      final data = res.data ?? {};
      final items = _parseItems(data);
      final total = (data['total'] as num?)?.toInt() ?? items.length;
      state = state.copyWith(items: items, page: 1, hasMore: items.length < total);
    } on ApiErrorException catch (e) {
      state = state.copyWith(errorMessage: () => e.message);
    }
  }

  Future<void> loadNextPage() async {
    if (state.isLoadingMore || !state.hasMore) return;
    state = state.copyWith(isLoadingMore: true);
    try {
      final nextPage = state.page + 1;
      final res = await _api.get<Map<String, dynamic>>(
        '/feed/home',
        queryParameters: {'page': nextPage, 'limit': _pageSize},
      );
      final data = res.data ?? {};
      final items = _parseItems(data);
      final total = (data['total'] as num?)?.toInt() ?? 0;
      final merged = [...state.items, ...items];
      state = state.copyWith(
        items: merged,
        page: nextPage,
        hasMore: merged.length < total,
        isLoadingMore: false,
      );
    } on ApiErrorException catch (e) {
      state = state.copyWith(isLoadingMore: false, errorMessage: () => e.message);
    }
  }

  List<HomeVideoItem> _parseItems(Map<String, dynamic> data) {
    final raw = data['items'];
    if (raw is! List) return [];
    return raw
        .whereType<Map<String, dynamic>>()
        .map(HomeVideoItem.fromApi)
        .where((item) => item.id.isNotEmpty)
        .toList();
  }

}

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - 280) {
      ref.read(homeFeedProvider.notifier).loadNextPage();
    }
  }

  @override
  void dispose() {
    _scrollController
      ..removeListener(_onScroll)
      ..dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final feed = ref.watch(homeFeedProvider);
    final items = feed.filteredItems;

    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(
                color: const Color(0xFFE11D48),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Icon(
                Icons.play_arrow_rounded,
                color: Colors.white,
                size: 18,
              ),
            ),
            const SizedBox(width: 10),
            const Text('GVP'),
          ],
        ),
        actions: [
          IconButton(
            tooltip: 'Go Live / Broadcast',
            icon: const Icon(Icons.sensors_rounded),
            onPressed: () => context.push('/live/live_creator_01?mode=broadcaster'),
          ),
          IconButton(
            tooltip: 'Notifications',
            icon: const Icon(Icons.notifications_none_rounded),
            onPressed: () => context.push('/notifications'),
          ),
          IconButton(
            tooltip: 'Search',
            icon: const Icon(Icons.search_rounded),
            onPressed: () => context.push('/search'),
          ),
          IconButton(
            tooltip: 'Settings',
            icon: const Icon(Icons.account_circle_outlined),
            onPressed: () => context.push('/settings'),
          ),
        ],
      ),
      body: Column(
        children: [
          SizedBox(
            height: 48,
            child: ListView.separated(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
              scrollDirection: Axis.horizontal,
              itemCount: HomeFeedNotifier.categories.length,
              separatorBuilder: (_, __) => const SizedBox(width: 8),
              itemBuilder: (context, index) {
                final category = HomeFeedNotifier.categories[index];
                final isSelected = feed.selectedCategory == category;
                return ChoiceChip(
                  label: Text(category),
                  selected: isSelected,
                  onSelected: (_) => ref
                      .read(homeFeedProvider.notifier)
                      .selectCategory(category),
                );
              },
            ),
          ),
          Expanded(
            child: Builder(
              builder: (context) {
                if (feed.isLoading) {
                  return const Center(child: CircularProgressIndicator());
                }
                if (feed.errorMessage != null && items.isEmpty) {
                  return Center(
                    child: Padding(
                      padding: const EdgeInsets.all(24),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(Icons.wifi_off_rounded, size: 48),
                          const SizedBox(height: 12),
                          Text(
                            feed.errorMessage!,
                            textAlign: TextAlign.center,
                          ),
                          const SizedBox(height: 16),
                          FilledButton(
                            onPressed: () => ref
                                .read(homeFeedProvider.notifier)
                                .refreshFeed(),
                            child: const Text('다시 시도'),
                          ),
                        ],
                      ),
                    ),
                  );
                }
                if (items.isEmpty) {
                  return const Center(
                    child: Text('표시할 영상이 없습니다.'),
                  );
                }
                return RefreshIndicator(
                  onRefresh: () =>
                      ref.read(homeFeedProvider.notifier).refreshFeed(),
                  child: ListView.builder(
                    key: const Key('home_feed_list'),
                    controller: _scrollController,
                    padding: const EdgeInsets.fromLTRB(12, 8, 12, 90),
                    itemCount: items.length + (feed.isLoadingMore ? 1 : 0),
                    itemBuilder: (context, index) {
                      if (index >= items.length) {
                        return const Padding(
                          padding: EdgeInsets.all(20),
                          child: Center(child: CircularProgressIndicator()),
                        );
                      }
                      final video = items[index];
                      return _VideoFeedCard(video: video);
                    },
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _VideoFeedCard extends StatelessWidget {
  final HomeVideoItem video;

  const _VideoFeedCard({required this.video});

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        key: Key('video_card_${video.id}'),
        onTap: () {
          if (video.isLive) {
            context.push('/live/${video.id}');
          } else {
            context.push(
              '/watch/${video.id}?title=${Uri.encodeComponent(video.title)}&channel=${Uri.encodeComponent(video.channelName)}',
            );
          }
        },
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AspectRatio(
              aspectRatio: 16 / 9,
              child: Stack(
                fit: StackFit.expand,
                children: [
                  if (video.thumbnailUrl != null &&
                      video.thumbnailUrl!.isNotEmpty)
                    Image.network(
                      video.thumbnailUrl!,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Container(
                        color: const Color(0xFF181826),
                        child: const Center(
                          child: Icon(
                            Icons.play_circle_outline_rounded,
                            size: 54,
                            color: Colors.white70,
                          ),
                        ),
                      ),
                    )
                  else
                    Container(
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          colors: [
                            video.accentColor.withValues(alpha: 0.45),
                            const Color(0xFF181826),
                          ],
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                        ),
                      ),
                      child: const Center(
                        child: Icon(
                          Icons.play_circle_outline_rounded,
                          size: 54,
                          color: Colors.white70,
                        ),
                      ),
                    ),
                  Positioned(
                    right: 10,
                    bottom: 10,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: video.isLive
                            ? const Color(0xFFE11D48)
                            : Colors.black87,
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Text(
                        video.durationLabel,
                        style: const TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(12),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  CircleAvatar(
                    radius: 18,
                    backgroundColor: video.accentColor,
                    child: Text(
                      video.channelName.substring(0, 1),
                      style: const TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          video.title,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            fontSize: 15,
                            fontWeight: FontWeight.w600,
                            height: 1.25,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '${video.channelName} • ${video.viewsText} • ${video.publishedText}',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade400,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
