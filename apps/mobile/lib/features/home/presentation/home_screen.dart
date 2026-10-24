import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

class HomeVideoItem {
  final String id;
  final String title;
  final String channelName;
  final String category;
  final String viewsText;
  final String publishedText;
  final String durationLabel;
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
    this.isLive = false,
    required this.accentColor,
  });
}

class HomeFeedState {
  final String selectedCategory;
  final List<HomeVideoItem> items;
  final bool isLoadingMore;
  final int page;

  const HomeFeedState({
    this.selectedCategory = 'All',
    this.items = const [],
    this.isLoadingMore = false,
    this.page = 1,
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
    bool? isLoadingMore,
    int? page,
  }) {
    return HomeFeedState(
      selectedCategory: selectedCategory ?? this.selectedCategory,
      items: items ?? this.items,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
      page: page ?? this.page,
    );
  }
}

final homeFeedProvider =
    StateNotifierProvider<HomeFeedNotifier, HomeFeedState>((ref) {
  return HomeFeedNotifier()..loadInitial();
});

class HomeFeedNotifier extends StateNotifier<HomeFeedState> {
  HomeFeedNotifier() : super(const HomeFeedState());

  static const List<String> categories = [
    'All',
    'Live',
    'Engineering',
    'Gaming',
    'Music',
    'Cinema 4K',
    'AI & Cloud',
  ];

  void loadInitial() {
    state = state.copyWith(items: _generateItemsForPage(1), page: 1);
  }

  void selectCategory(String category) {
    state = state.copyWith(selectedCategory: category);
  }

  Future<void> refreshFeed() async {
    await Future<void>.delayed(const Duration(milliseconds: 250));
    state = state.copyWith(items: _generateItemsForPage(1), page: 1);
  }

  Future<void> loadNextPage() async {
    if (state.isLoadingMore) return;
    state = state.copyWith(isLoadingMore: true);
    await Future<void>.delayed(const Duration(milliseconds: 200));
    final nextPage = state.page + 1;
    state = state.copyWith(
      items: [...state.items, ..._generateItemsForPage(nextPage)],
      page: nextPage,
      isLoadingMore: false,
    );
  }

  List<HomeVideoItem> _generateItemsForPage(int page) {
    final base = (page - 1) * 4;
    return [
      HomeVideoItem(
        id: 'vid_home_${base + 1}',
        title:
            'Global Video Mesh: Sub-Second LL-HLS & AV1 Edge Transcoding (#${base + 1})',
        channelName: 'GVP Engineering',
        category: 'Engineering',
        viewsText: '284K views',
        publishedText: '3 hours ago',
        durationLabel: '24:18',
        accentColor: const Color(0xFFE11D48),
      ),
      HomeVideoItem(
        id: 'live_home_${base + 2}',
        title: 'LIVE: Global Creator Summit 2026 - 4K60 Keynote Broadcast',
        channelName: 'GVP Official Live',
        category: 'Live',
        viewsText: '19.4K watching',
        publishedText: 'STREAMING NOW',
        durationLabel: 'LIVE',
        isLive: true,
        accentColor: const Color(0xFFEF4444),
      ),
      HomeVideoItem(
        id: 'vid_home_${base + 3}',
        title: 'Patagonia Dolby Vision 4K 60fps HDR Mastering Showcase',
        channelName: 'Cinema4K HDR Lab',
        category: 'Cinema 4K',
        viewsText: '1.2M views',
        publishedText: '2 days ago',
        durationLabel: '18:05',
        accentColor: const Color(0xFF6366F1),
      ),
      HomeVideoItem(
        id: 'vid_home_${base + 4}',
        title:
            'Distributed ClickHouse Analytics: Real-Time QoE Telemetry at Scale',
        channelName: 'Cloud Native Foundry',
        category: 'AI & Cloud',
        viewsText: '92K views',
        publishedText: '5 days ago',
        durationLabel: '31:42',
        accentColor: const Color(0xFF0EA5E9),
      ),
    ];
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
            child: RefreshIndicator(
              onRefresh: () => ref.read(homeFeedProvider.notifier).refreshFeed(),
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
